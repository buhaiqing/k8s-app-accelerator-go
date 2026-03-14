package generator

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/template"
)

// init 初始化随机数种子
func init() {
	rand.Seed(time.Now().UnixNano())
}

// CMDBGenerator CMDB SQL 生成器
type CMDBGenerator struct {
	config      *config.ProjectConfig
	resources   *config.ResourceGroup
	outputDir   string
	templateDir string
	workerPool  *template.WorkerPool
}

// NewCMDBGenerator 创建新的 CMDB 生成器
func NewCMDBGenerator(
	cfg *config.ProjectConfig,
	resources *config.ResourceGroup,
	outputDir string,
	templateDir string,
	scriptPath string,
) (*CMDBGenerator, error) {
	// 创建 Worker 池（5 个 workers）
	pool, err := template.NewWorkerPool(5, scriptPath)
	if err != nil {
		return nil, fmt.Errorf("创建 Worker 池失败：%w", err)
	}

	return &CMDBGenerator{
		config:      cfg,
		resources:   resources,
		outputDir:   outputDir,
		templateDir: templateDir,
		workerPool:  pool,
	}, nil
}

// Close 关闭资源
func (g *CMDBGenerator) Close() {
	if g.workerPool != nil {
		g.workerPool.Close()
	}
}

// GenerateAll 生成所有 SQL
func (g *CMDBGenerator) GenerateAll() error {
	fmt.Printf("开始生成 CMDB SQL 脚本...\n")

	// 解析资源变量
	resourceVars, err := g.resolveResources()
	if err != nil {
		return fmt.Errorf("解析资源失败：%w", err)
	}

	// 为每个profile 生成 SQL
	for _, profile := range g.config.Profiles {
		if err := g.generateForProfile(profile, resourceVars); err != nil {
			return err
		}
	}

	fmt.Printf("✓ 成功生成 CMDB SQL 脚本\n")
	return nil
}

// generateForProfile 为特定环境生成 SQL
func (g *CMDBGenerator) generateForProfile(profile string, resourceVars map[string]interface{}) error {
	fmt.Printf("  [INFO] 生成 profile: %s\n", profile)

	// 构建 stack 列表（去重）
	stackValues := make([]string, 0, len(g.config.Stack))
	for _, stackId := range g.config.Stack {
		if !contains(stackValues, stackId) {
			stackValues = append(stackValues, stackId)
		}
	}

	// 准备渲染上下文（不包含 stackid）
	baseContext := make(map[string]interface{})
	baseContext["profile"] = profile
	baseContext["stack"] = g.config.Stack // dict 形式

	// 合并资源变量到 baseContext
	for k, v := range resourceVars {
		baseContext[k] = v
	}

	// 生成 Ansible 兼容的动态 ID（与 generate.yml 一致）
	baseContext["rds_db"] = fmt.Sprintf("%s-rds-%d", profile, randomInt(20000))
	baseContext["pg_db"] = fmt.Sprintf("%s-pg-%d", profile, randomInt(20000))
	baseContext["_mongo_db"] = fmt.Sprintf("%s-mongo-%d", profile, randomInt(20000))

	// 查找模板文件
	sqlTemplate := filepath.Join(g.templateDir, "sql.j2")
	if _, err := os.Stat(sqlTemplate); os.IsNotExist(err) {
		return fmt.Errorf("SQL 模板不存在：%s", sqlTemplate)
	}

	// 为每个 stack 生成 SQL 片段
	var allSQLParts []string
	for i, stackId := range stackValues {
		// 为每个 stack 创建独立的上下文
		ctx := make(map[string]interface{})
		for k, v := range baseContext {
			ctx[k] = v
		}
		ctx["stackid"] = stackId

		// 生成随机值（模拟 Ansible的 random filter）
		ctx["random_value"] = generateRandomString(5)

		// 渲染模板
		content, err := g.renderTemplate(sqlTemplate, ctx)
		if err != nil {
			return fmt.Errorf("渲染 SQL 模板失败（stack=%s）：%w", stackId, err)
		}

		// 只保留第一个 stack 的 dockerhub INSERT（避免重复）
		if i > 0 {
			// 移除 dockerhub INSERT 行
			lines := strings.Split(content, "\n")
			filteredLines := make([]string, 0)
			for _, line := range lines {
				if !strings.Contains(line, "INSERT INTO `dockerhub`") {
					filteredLines = append(filteredLines, line)
				}
			}
			content = strings.Join(filteredLines, "\n")
		}

		allSQLParts = append(allSQLParts, content)
	}

	// 合并所有 SQL 片段
	finalSQL := strings.Join(allSQLParts, "\n\n")

	// 写入 SQL 文件
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败：%w", err)
	}
	outputFile := filepath.Join(g.outputDir, fmt.Sprintf("sql_%s.sql", profile))
	if err := os.WriteFile(outputFile, []byte(finalSQL), 0644); err != nil {
		return fmt.Errorf("写入 SQL 文件失败：%w", err)
	}

	fmt.Printf("    - 已写入：%s (共 %d 个 stacks)\n", outputFile, len(stackValues))

	// 复制 initsql 文件
	initsqlSrc := filepath.Join(g.templateDir, "inittables.sql")
	if _, err := os.Stat(initsqlSrc); err == nil {
		initsqlDest := filepath.Join(g.outputDir, "inittables.sql")
		if err := copyFile(initsqlSrc, initsqlDest); err != nil {
			return fmt.Errorf("复制 initsql 文件失败：%w", err)
		}
		fmt.Printf("    - 已复制：inittables.sql\n")
	}

	return nil
}

// contains 检查 slice 是否包含某元素
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// resolveResources 解析资源变量
func (g *CMDBGenerator) resolveResources() (map[string]interface{}, error) {
	vars := make(map[string]interface{})

	// 解析 RDS 资源
	if len(g.resources.RDS) > 0 {
		rds := g.resources.RDS[0]
		vars["datasource_url"] = rds.DatasourceURL
		vars["datasource_port"] = rds.DatasourcePort
		vars["datasource_db"] = rds.DatasourceDB
		vars["datasource_username"] = rds.DatasourceUsername
		vars["datasource_password"] = rds.DatasourcePassword
		vars["rds_db"] = fmt.Sprintf("rds_%s", rds.Name)
	}

	// 解析 PostgreSQL 资源
	if len(g.resources.PostgreSQL) > 0 {
		pg := g.resources.PostgreSQL[0]
		vars["datasource_pg_url"] = pg.DatasourcePgURL
		vars["datasource_pg_port"] = pg.DatasourcePgPort
		vars["datasource_pg_db"] = pg.DatasourcePgDB
		vars["datasource_pg_username"] = pg.DatasourcePgUsername
		vars["datasource_pg_password"] = pg.DatasourcePgPassword
		vars["pg_db"] = fmt.Sprintf("pg_%s", pg.Name)
	}

	// 解析 Redis 资源
	if len(g.resources.Redis) > 0 {
		redis := g.resources.Redis[0]
		vars["redisIp"] = redis.RedisIP
		vars["redisPort"] = redis.RedisPort
		vars["redisDb"] = redis.RedisDb
		vars["redisPassword"] = redis.RedisPassword
	}

	// 解析 MongoDB 资源
	if len(g.resources.MongoDB) > 0 {
		mongo := g.resources.MongoDB[0]
		vars["mongo_db"] = mongo.MongoDB
		vars["mongodb_url"] = mongo.MongoDBURL
		vars["mongodb_username"] = mongo.MongoDBUsername
		vars["mongodb_password"] = mongo.MongoDBPassword
		vars["_mongo_db"] = fmt.Sprintf("mongo_%s", mongo.Name)
	}

	return vars, nil
}

// renderTemplate 渲染模板
func (g *CMDBGenerator) renderTemplate(templatePath string, context map[string]interface{}) (string, error) {
	// 从 Worker 池获取 worker
	worker := g.workerPool.GetWorker()

	// 调用 Worker 渲染
	req := template.RenderRequest{
		TemplatePath: templatePath,
		Context:      context,
	}
	result, err := worker.Render(req)
	if err != nil {
		return "", fmt.Errorf("渲染模板失败：%w", err)
	}

	return result, nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// generateRandomString 生成随机字符串
func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

// randomInt 生成随机整数（模拟 Ansible的 random filter）
func randomInt(max int) int {
	return rand.Intn(max)
}
