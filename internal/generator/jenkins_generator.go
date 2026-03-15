package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/template"
)

// JenkinsGenerator Jenkins Jobs 配置生成器
type JenkinsGenerator struct {
	projectConfig *config.ProjectConfig
	outputDir     string
	templateDir   string
	workerPool    *template.WorkerPool
}

// JenkinsJobData Jenkins Job 数据结构（对应 Ansible 中的 item）
type JenkinsJobData struct {
	Common       CommonConfig `yaml:",inline"`
	DNETProduct  string       `yaml:"DNET_PRODUCT"`
	ProductDes   string       `yaml:"product_des"`
	Output       string       `yaml:"output"`
	Receivers    string       `yaml:"receivers"`
	Env          string       `yaml:"env"`
	Surfix       string       `yaml:"surfix"`
	GitBaseURL   string       `yaml:"GIT_BASE_URL"`
	GitBaseGroup string       `yaml:"GIT_BASE_GROUP"`
	DNETProject  string       `yaml:"DNET_PROJECT"`
}

// CommonConfig 通用配置
type CommonConfig struct {
	DNETProject  string `yaml:"DNET_PROJECT"`
	GitBaseURL   string `yaml:"GIT_BASE_URL"`
	GitBaseGroup string `yaml:"GIT_BASE_GROUP"`
	Output       string `yaml:"output"`
	Receivers    string `yaml:"receivers"`
	Env          string `yaml:"env"`
	Surfix       string `yaml:"surfix"`
}

// NewJenkinsGenerator 创建新的 Jenkins 生成器
func NewJenkinsGenerator(
	projectConfig *config.ProjectConfig,
	outputDir string,
	templateDir string,
	scriptPath string,
) (*JenkinsGenerator, error) {
	pool, err := template.NewWorkerPool(5, scriptPath)
	if err != nil {
		return nil, fmt.Errorf("创建 worker 池失败：%w", err)
	}

	return &JenkinsGenerator{
		projectConfig: projectConfig,
		outputDir:     outputDir,
		templateDir:   templateDir,
		workerPool:    pool,
	}, nil
}

// Close 关闭资源
func (g *JenkinsGenerator) Close() {
	if g.workerPool != nil {
		g.workerPool.Close()
	}
}

// GenerateAll 生成所有 Jenkins Jobs 配置
func (g *JenkinsGenerator) GenerateAll() error {
	fmt.Printf("开始生成 Jenkins Jobs 配置...\n")

	// 从配置中读取 data 列表
	// 这里我们需要加载 vars.yaml 中的 data 字段
	dataList, err := g.loadJenkinsData()
	if err != nil {
		return fmt.Errorf("加载 Jenkins 数据失败：%w", err)
	}

	fmt.Printf("找到 %d 个 Jenkins Job 配置\n", len(dataList))

	// 为每个配置生成 Jenkins Job
	for _, data := range dataList {
		if err := g.GenerateForProduct(data); err != nil {
			return fmt.Errorf("生成 %s 失败：%w", data.DNETProduct, err)
		}
	}

	fmt.Printf("✓ 成功生成 %d 个 Jenkins Jobs\n", len(dataList))
	return nil
}

// GenerateForProduct 为单个产品生成 Jenkins Job 配置
func (g *JenkinsGenerator) GenerateForProduct(data JenkinsJobData) error {
	// 计算名称和 folder
	name := fmt.Sprintf("%s_%s_k8s", data.DNETProduct, data.Surfix)
	folder := fmt.Sprintf("%s_%s_K8s", strings.ToUpper(data.DNETProduct), data.Surfix)
	prefix := strings.ToUpper(data.DNETProduct)

	fmt.Printf("  - 生成 %s (%s)\n", name, data.ProductDes)

	// 构建渲染上下文
	ctx := map[string]interface{}{
		"name":           name,
		"folder":         folder,
		"prefix":         prefix,
		"suffix":         data.Surfix,
		"environment":    data.Env,
		"product_des":    "{customer} " + data.ProductDes,
		"receivers":      data.Receivers,
		"DNET_PROFILE":   []string{"production"},
		"git_branches":   []string{"master"},
		"GIT_BRANCH":     "develop",
		"check_rdb":      []string{"no", "yes"},
		"check_config":   []string{"no", "yes"},
		"jenkinsfiletop": "kubernetes",
		"DNET_PROJECT":   data.DNETProject,
		"DNET_PRODUCT":   data.DNETProduct,
		"credentials_id": "{credentials_id}",
		"GIT_BASE_URL":   data.GitBaseURL,
		"GIT_BASE_GROUP": data.GitBaseGroup,
		"on_k8s":         true,
		"item": map[string]interface{}{
			"receivers":      data.Receivers,
			"surfix":         data.Surfix,
			"env":            data.Env,
			"product_des":    data.ProductDes,
			"DNET_PROJECT":   data.DNETProject,
			"DNET_PRODUCT":   data.DNETProduct,
			"GIT_BASE_URL":   data.GitBaseURL,
			"GIT_BASE_GROUP": data.GitBaseGroup,
		},
		"env_vars": fmt.Sprintf(`DNET_PROJECT=%s
DNET_PRODUCT=%s
credentials_id={credentials_id}
GIT_BASE_URL=%s
GIT_BASE_GROUP=%s
on_k8s=True`, data.DNETProject, data.DNETProduct, data.GitBaseURL, data.GitBaseGroup),
	}

	// 使用 Python Worker 渲染模板
	// 模板路径：{templateDir}/jobs/templates/job.j2
	// 其中 templateDir 指向 Ansible roles 目录
	templatePath := filepath.Join(g.templateDir, "jobs", "templates", "job.j2")

	// 转换为绝对路径
	absTemplatePath, err := filepath.Abs(templatePath)
	if err != nil {
		return fmt.Errorf("获取模板绝对路径失败：%w", err)
	}

	fmt.Printf("    模板路径：%s\n", absTemplatePath)

	content, err := g.workerPool.Render(template.RenderRequest{
		TemplatePath: absTemplatePath,
		Context:      ctx,
	})
	if err != nil {
		return fmt.Errorf("渲染模板失败：%w", err)
	}

	// 写入文件：{outputDir}/{DNET_PRODUCT}/project.yml
	outputPath := filepath.Join(g.outputDir, data.DNETProduct, "project.yml")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败：%w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入文件失败：%w", err)
	}

	fmt.Printf("    ✅ 已生成：%s\n", outputPath)
	return nil
}

// loadJenkinsData 加载 Jenkins Job 数据
// 从 projectConfig 中解析 data 列表
func (g *JenkinsGenerator) loadJenkinsData() ([]JenkinsJobData, error) {
	var dataList []JenkinsJobData

	// 从 projectConfig 中读取 data 字段
	for _, data := range g.projectConfig.Data {
		// 合并 common 配置到每个 data 项
		jobData := JenkinsJobData{
			Common: CommonConfig{
				DNETProject:  data.Common.DNETProject,
				GitBaseURL:   data.Common.GitBaseURL,
				GitBaseGroup: data.Common.GitBaseGroup,
				Output:       data.Common.Output,
				Receivers:    data.Common.Receivers,
				Env:          data.Common.Env,
				Surfix:       data.Common.Surfix,
			},
			DNETProduct:  data.DNETProduct,
			ProductDes:   data.ProductDes,
			Output:       g.outputDir,
			Receivers:    data.Common.Receivers,
			Env:          data.Common.Env,
			Surfix:       data.Common.Surfix,
			GitBaseURL:   data.Common.GitBaseURL,
			GitBaseGroup: data.Common.GitBaseGroup,
			DNETProject:  data.Common.DNETProject,
		}
		dataList = append(dataList, jobData)
	}

	return dataList, nil
}
