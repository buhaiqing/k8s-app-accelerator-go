package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/model"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/template"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// Generator 配置生成器
type Generator struct {
	config      *config.ProjectConfig
	resources   *config.ResourceGroup
	mapping     *config.Mapping
	roleVars    []*model.RoleVars
	outputDir   string
	templateDir string
	workerPool  *template.WorkerPool
}

// NewGenerator 创建新的生成器
func NewGenerator(
	cfg *config.ProjectConfig,
	resources *config.ResourceGroup,
	mapping *config.Mapping,
	roleVars []*model.RoleVars,
	outputDir string,
	templateDir string,
) (*Generator, error) {
	return NewGeneratorWithScript(cfg, resources, mapping, roleVars, outputDir, templateDir, "scripts/render_worker.py")
}

// NewGeneratorWithScript 创建新的生成器（指定脚本路径）
func NewGeneratorWithScript(
	cfg *config.ProjectConfig,
	resources *config.ResourceGroup,
	mapping *config.Mapping,
	roleVars []*model.RoleVars,
	outputDir string,
	templateDir string,
	scriptPath string,
) (*Generator, error) {
	// 创建 Worker 池（5 个 workers）
	pool, err := template.NewWorkerPool(5, scriptPath)
	if err != nil {
		return nil, fmt.Errorf("创建 Worker 池失败：%w", err)
	}

	return &Generator{
		config:      cfg,
		resources:   resources,
		mapping:     mapping,
		roleVars:    roleVars,
		outputDir:   outputDir,
		templateDir: templateDir,
		workerPool:  pool,
	}, nil
}

// Close 关闭资源
func (g *Generator) Close() {
	if g.workerPool != nil {
		g.workerPool.Close()
	}
}

// GenerateAll 生成所有配置
func (g *Generator) GenerateAll() error {
	return g.GenerateAllWithContext(context.Background())
}

// GenerateAllWithContext 生成所有配置（支持上下文取消）
func (gen *Generator) GenerateAllWithContext(ctx context.Context) error {
	fmt.Printf("开始生成配置，共 %d 个应用...\n", len(gen.roleVars))

	// 使用 errgroup 管理并发和错误
	// 设置并发限制为 10（避免资源耗尽）
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(10)

	// 收集所有错误
	var errors []error
	var errorMu sync.Mutex

	for _, roleVars := range gen.roleVars {
		rv := roleVars // 捕获变量

		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := gen.generateForApp(rv); err != nil {
					err := fmt.Errorf("生成 %s 配置失败：%w", rv.App, err)
					errorMu.Lock()
					errors = append(errors, err)
					errorMu.Unlock()
					return err
				}
				return nil
			}
		})
	}

	// 等待所有任务完成
	if err := eg.Wait(); err != nil {
		// 返回所有错误的汇总
		if len(errors) > 0 {
			return fmt.Errorf("生成配置失败（共 %d 个错误）：%v", len(errors), errors[0])
		}
		return err
	}

	fmt.Printf("✓ 成功生成 %d 个应用的配置\n", len(gen.roleVars))
	return nil
}

// generateForApp 为单个应用生成配置
func (g *Generator) generateForApp(roleVars *model.RoleVars) error {
	appName := roleVars.App

	// 获取 product
	product, exists := g.mapping.Mappings[appName]
	if !exists {
		return fmt.Errorf("在 mapping.yaml 中未找到 %s 的映射", appName)
	}

	// 为每个 profile 生成配置
	for _, profile := range g.config.Profiles {
		if err := g.generateForProfile(appName, product, profile, roleVars); err != nil {
			return err
		}
	}

	return nil
}

// generateForProfile 为特定环境生成配置
func (g *Generator) generateForProfile(appName, product, profile string, roleVars *model.RoleVars) error {
	// 获取 stackId
	stackId, exists := g.config.Stack[appName]
	if !exists {
		stackId = profile // 默认使用 profile
	}

	// 构建输出目录结构：{outputDir}/{product}/{appName}/overlays/{profile}/{stackId}
	// 这与 Ansible 的目录结构完全一致
	outputPath := filepath.Join(g.outputDir, product, appName, "overlays", profile, stackId)

	fmt.Printf("  [INFO] 生成 %s (%s) - profile: %s, stackId: %s\n",
		appName, product, profile, stackId)

	// 创建目录
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("创建目录失败：%w", err)
	}

	// 解析资源变量
	resourceVars, err := g.resolveResources()
	if err != nil {
		return fmt.Errorf("解析资源失败：%w", err)
	}

	// 生成随机值（用于 Job 名称唯一性）
	randomValue := generateRandomValue()

	// 生成 Jenkins 定义
	jenkinsDefs := g.generateJenkinsDefs(product, profile)

	// 确定 harbor_project
	// 优先级: roleVars.HarborProject > mapping.Mappings[appName] > g.config.Project
	harborProject := roleVars.HarborProject
	if harborProject == "" {
		harborProject = product // product 就是映射中的值
	}

	// 准备渲染上下文（与 Ansible 完全兼容）
	context := map[string]interface{}{
		// 基础变量
		"app":     appName,
		"profile": profile,
		"product": product,
		"project": g.config.Project,
		"_type":   roleVars.Type,
		"stackId": stackId,

		// Namespace（使用 stackId 或 profile）
		"namespace": stackId,

		// HPA 配置（从 role vars 读取，如果没有则使用默认值）
		"minReplicas": 2,
		"maxReplicas": 2, // 默认值与 Ansible 一致

		// Apollo 配置
		"apollo": map[string]interface{}{
			"site":       g.config.Apollo.Site,
			"customerid": g.config.Apollo.CustomerID,
			"env":        g.config.Apollo.Env,
			"alias":      g.config.Apollo.Alias,
			"token":      g.config.Apollo.Token,
		},

		// ArgoCD 配置
		"argocd": map[string]interface{}{
			"site":    g.config.ArgoCD.Site,
			"cluster": g.config.ArgoCD.Cluster,
		},

		// Jenkins 配置
		"jenkins": map[string]interface{}{
			"site": g.config.Jenkins.Site,
		},
		"jenkins_defs": jenkinsDefs,

		// Harbor 项目
		"harbor_project": harborProject,

		// 资源配置（结构化）
		"resources": roleVars.Resources,

		// 映射字典
		"mappings": g.mapping.Mappings,

		// 随机值
		"random_value": randomValue,

		// DB 迁移配置（从 role vars 读取）
		// 如果 roleVars.SetupImage 为空，则使用默认格式
		"setup_image":       getSetupImage(roleVars, harborProject),
		"setup_db":          roleVars.SetupDB,
		"setup_db_fallback": appName, // 如果 setup_db 为空则使用 app 名称

		// DNET_PRODUCT（用于 Jenkins job 命名）
		"DNET_PRODUCT": product,

		// 版本号
		"version": "1.0",

		// 资源变量（扁平化）
		// 从 resourceVars 合并到 context
	}

	// 如果 setup_db 为空,使用 appName
	if context["setup_db"] == "" {
		context["setup_db"] = appName
	}

	// 如果 setup_image 为空，生成默认值
	if context["setup_image"] == "" {
		context["setup_image"] = fmt.Sprintf("harbor.qianfan123.com/%s/%s-rdb-setup", harborProject, appName)
	}

	// 合并资源变量到 context
	for k, v := range resourceVars {
		context[k] = v
	}

	// 获取模板文件列表
	templateFiles, err := g.getTemplateFiles(product, profile, roleVars)
	if err != nil {
		return fmt.Errorf("获取模板文件失败：%w", err)
	}

	fmt.Printf("  [INFO] 找到 %d 个模板文件\n", len(templateFiles))

	// 渲染每个模板文件
	for _, tmplFile := range templateFiles {
		fmt.Printf("    - 渲染：%s\n", filepath.Base(tmplFile))
		if err := g.renderTemplate(tmplFile, outputPath, context); err != nil {
			return fmt.Errorf("渲染 %s 失败：%w", tmplFile, err)
		}
	}

	return nil
}

// resolveResources 解析资源变量（从 resources.yaml 提取扁平化变量）
// 与 Ansible 的变量命名保持完全一致
func (g *Generator) resolveResources() (map[string]interface{}, error) {
	vars := make(map[string]interface{})

	// 解析 RDS 资源
	if len(g.resources.RDS) > 0 {
		rds := g.resources.RDS[0] // 使用 default 资源组
		vars["datasource_url"] = rds.DatasourceURL
		vars["datasource_port"] = rds.DatasourcePort
		vars["datasource_db"] = rds.DatasourceDB
		vars["datasource_username"] = rds.DatasourceUsername
		vars["datasource_password"] = rds.DatasourcePassword
	}

	// 解析 PostgreSQL 资源
	if len(g.resources.PostgreSQL) > 0 {
		pg := g.resources.PostgreSQL[0]
		vars["datasource_pg_url"] = pg.DatasourcePgURL
		vars["datasource_pg_port"] = pg.DatasourcePgPort
		vars["datasource_pg_db"] = pg.DatasourcePgDB
		vars["datasource_pg_username"] = pg.DatasourcePgUsername
		vars["datasource_pg_password"] = pg.DatasourcePgPassword
	}

	// 解析 Oracle 资源
	if len(g.resources.Oracle) > 0 {
		oracle := g.resources.Oracle[0]
		vars["oracle_url"] = oracle.OracleURL
		vars["oracle_port"] = oracle.OraclePort
		vars["oracle_db"] = oracle.OracleDB
		vars["oracle_user"] = oracle.OracleUser
		vars["oracle_pass"] = oracle.OraclePass
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
	}

	// 解析 Elasticsearch 资源
	if len(g.resources.Elasticsearch) > 0 {
		es := g.resources.Elasticsearch[0]
		vars["es_instanceId"] = es.ESInstanceID
		vars["es_domain"] = es.ESDomain
		vars["es_secret"] = es.ESSecret
		vars["es_regionId"] = es.ESRegionID
		vars["es_accessKeyId"] = es.ESAccessKeyID
		vars["es_url"] = es.ESURL
		vars["es_username"] = es.ESUsername
		vars["es_password"] = es.ESPassword
	}

	// 解析 OSS 资源
	if len(g.resources.OSS) > 0 {
		oss := g.resources.OSS[0]
		vars["oss_roleArn"] = oss.OSSRoleArn
		vars["oss_stsEndpoint"] = oss.OSSStsEndpoint
		vars["bucketName"] = oss.BucketName
		vars["ossEndpoint"] = oss.OSSEndpoint
		vars["ossimageBaseUrl"] = oss.OSSImageBaseURL
		vars["ossinternalEndpoint"] = oss.OSSInternalEndpoint
		vars["oss_accessKeyId"] = oss.OSSAccessKeyID
		vars["oss_accessKeySecret"] = oss.OSSAccessKeySecret
	}

	// 解析 MQ 资源
	if len(g.resources.MQ) > 0 {
		mq := g.resources.MQ[0]
		vars["cmq_secretId"] = mq.CMQSecretID
		vars["cmq_secretKey"] = mq.CMQSecretKey
		vars["cmq_queueEndpoint"] = mq.CMQQueueEndpoint
		vars["cmq_topicEndpoint"] = mq.CMQTopicEndpoint
		vars["mns_accessKeyId"] = mq.MNSAccessKeyID
		vars["mns_accessKeySecret"] = mq.MNSAccessKeySecret
		vars["mns_endpoint"] = mq.MNSEndpoint
		vars["mns_stuffix"] = mq.MNSStuffix
		vars["rocketmq_accesskey"] = mq.RocketMQAccessKey
		vars["rocketmq_secretkey"] = mq.RocketMQSecretKey
		vars["rocketmq_instanceId"] = mq.RocketMQInstanceID
		vars["rocketmq_regionId"] = mq.RocketMQRegionID
		vars["rocketmq_namesrvaddr"] = mq.RocketMQNamesrvaddr
	}

	// 解析 RabbitMQ 资源
	if len(g.resources.RabbitMQ) > 0 {
		rb := g.resources.RabbitMQ[0]
		vars["rabbitmq_host"] = rb.Host
		vars["rabbitmq_user"] = rb.User
		vars["rabbitmq_password"] = rb.Password
		vars["rabbitmq_port"] = fmt.Sprintf("%d", rb.Port)
		vars["rabbitmq_retry_delay_time"] = fmt.Sprintf("%d", rb.RetryDelayTime)
	}

	// 解析 OTS 资源
	if len(g.resources.OTS) > 0 {
		ots := g.resources.OTS[0]
		vars["ots_endPoint"] = ots.OTSEndPoint
		vars["ots_accessKeyId"] = ots.OTSAccessKeyID
		vars["ots_accessKeySecret"] = ots.OTSAccessKeySecret
		vars["ots_instanceName"] = ots.OTSInstanceName
	}

	// 解析 DTFlow 资源
	if len(g.resources.DTFlow) > 0 {
		dt := g.resources.DTFlow[0]
		vars["dft_config_serverUrl"] = dt.DFTConfigServerURL
		vars["dft_config_tenant"] = dt.DFTConfigTenant
		vars["dft_config_username"] = dt.DFTConfigUsername
		vars["dft_config_password"] = dt.DFTConfigPassword
	}

	return vars, nil
}

// generateJenkinsDefs 生成 Jenkins job 定义
// 与 Ansible 的格式保持一致：product 需要大写
func (g *Generator) generateJenkinsDefs(product, profile string) map[string]interface{} {
	// 使用 profile_convert 逻辑
	profileConverted := convertProfile(profile)
	// product 需要大写（与 Ansible 的 DNET_PRODUCT|upper 一致）
	productUpper := strings.ToUpper(product)

	return map[string]interface{}{
		"appinstall":     fmt.Sprintf("%s_%s_K8s/%s_deploy_no_rdb_%s", productUpper, profileConverted, productUpper, profileConverted),
		"rdb_appinstall": fmt.Sprintf("%s_%s_K8s/%s_deploy_%s", productUpper, profileConverted, productUpper, profileConverted),
		"rdb":            fmt.Sprintf("%s_%s_K8s/%s_deploy_rdb_%s", productUpper, profileConverted, productUpper, profileConverted),
	}
}

// convertProfile 转换 profile 名称（与 Ansible filter_plugins/profile.py 一致）
func convertProfile(profile string) string {
	switch profile {
	case "int":
		return "Int"
	case "uat":
		return "Uat"
	case "branch":
		return "BRA"
	case "production":
		return "PRD"
	default:
		return profile
	}
}

// generateRandomValue 生成随机值（用于 Job 名称唯一性）
func generateRandomValue() string {
	// 模拟 Ansible: {{ 10000 | random | to_uuid | truncate(7,True,'') }}
	u := uuid.New()
	return u.String()[:7]
}

// getTemplateFiles 获取模板文件列表（支持条件生成）
// 模板路径与 Ansible 一致：overlays/production/production/{template}.j2
func (g *Generator) getTemplateFiles(product, profile string, roleVars *model.RoleVars) ([]string, error) {
	// 模板目录：{templateDir}/{product}/templates/overlays/production/production
	// 这与 Ansible 的 overlays.yml 中的路径逻辑一致
	// src: "overlays/{{_path}}/{{_path}}/{{config_file}}.j2" 其中 _path = "production"
	templateDir := filepath.Join(g.templateDir, product, "templates", "overlays", "production", "production")

	fmt.Printf("  [DEBUG] 模板目录：%s\n", templateDir)

	files := make([]string, 0)

	// 总是生成的模板（与 Ansible overlays.yml 一致）
	alwaysTemplates := []string{
		"cm-suffix-transformer.yaml.j2",
		"config.yaml.j2",
		"deployment.yaml.j2",
		"dev.env.j2",
		"kustomization.yaml.j2",
		"ops.env.j2",
		"ops.j2.j2",
		"service.yaml.j2",
	}

	for _, tmpl := range alwaysTemplates {
		path := filepath.Join(templateDir, tmpl)
		if _, err := os.Stat(path); err == nil {
			files = append(files, path)
		} else {
			fmt.Printf("  [DEBUG] 模板不存在：%s\n", path)
		}
	}

	// 条件生成 hpa.yaml（当 enable_hpa == true）
	if roleVars.EnableHPA {
		hpaPath := filepath.Join(templateDir, "hpa.yaml.j2")
		if _, err := os.Stat(hpaPath); err == nil {
			files = append(files, hpaPath)
		}
	}

	// 条件生成 job.yaml（当 enable_rdb == true 且 _type == "backend"）
	if roleVars.EnableRDB && roleVars.Type == "backend" {
		jobPath := filepath.Join(templateDir, "job.yaml.j2")
		if _, err := os.Stat(jobPath); err == nil {
			files = append(files, jobPath)
		}
	}

	return files, nil
}

// renderTemplate 渲染单个模板文件
func (g *Generator) renderTemplate(templatePath, outputPath string, context map[string]interface{}) error {
	// 从 Worker 池获取 worker
	worker := g.workerPool.GetWorker()

	// 调用 Worker 渲染
	req := template.RenderRequest{
		TemplatePath: templatePath,
		Context:      context,
	}
	result, err := worker.Render(req)
	if err != nil {
		return fmt.Errorf("渲染模板失败：%w", err)
	}

	// 确定输出文件名
	baseName := filepath.Base(templatePath)
	var outputFile string
	if len(baseName) > 3 && baseName[len(baseName)-3:] == ".j2" {
		// 移除 .j2 后缀
		outputFile = filepath.Join(outputPath, baseName[:len(baseName)-3])
	} else {
		outputFile = filepath.Join(outputPath, baseName)
	}

	// 写入文件（确保末尾有换行符，与 Ansible 保持一致）
	content := result
	if len(content) > 0 && content[len(content)-1] != '\n' {
		content += "\n"
	}
	if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入文件失败：%w", err)
	}

	fmt.Printf("    [INFO] 已写入：%s\n", outputFile)
	return nil
}

// GenerateIncremental 增量生成（只生成变更的部分）
func (g *Generator) GenerateIncremental(changedApps []string) error {
	fmt.Printf("开始增量生成，共 %d 个应用变更...\n", len(changedApps))

	for _, appName := range changedApps {
		// 找到对应的 roleVars
		var roleVars *model.RoleVars
		for _, rv := range g.roleVars {
			if rv.App == appName {
				roleVars = rv
				break
			}
		}

		if roleVars == nil {
			return fmt.Errorf("未找到应用 %s 的配置", appName)
		}

		if err := g.generateForApp(roleVars); err != nil {
			return err
		}
	}

	fmt.Printf("✓ 成功生成 %d 个应用的配置\n", len(changedApps))
	return nil
}

// getSetupImage 获取 setup_image 变量
// 优先从 role vars 读取，如果没有则使用默认值
func getSetupImage(roleVars *model.RoleVars, harborProject string) string {
	if roleVars.SetupImage != "" {
		return roleVars.SetupImage
	}
	// 默认值：harbor.qianfan123.com/{harbor_project}/{app}-rdb-setup
	return fmt.Sprintf("harbor.qianfan123.com/%s/%s-rdb-setup", harborProject, roleVars.App)
}
