package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/model"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/template"
)

// ArgoCDGenerator ArgoCD Application 生成器
type ArgoCDGenerator struct {
	projectConfig *config.ProjectConfig
	roleVars      []*model.RoleVars
	templateDir   string
	outputDir     string
	workerPool    *template.WorkerPool
}

// NewArgoCDGenerator 创建新的生成器
func NewArgoCDGenerator(
	projectConfig *config.ProjectConfig,
	roleVars []*model.RoleVars,
	outputDir string,
	templateDir string,
	scriptPath string,
) (*ArgoCDGenerator, error) {
	pool, err := template.NewWorkerPool(5, scriptPath)
	if err != nil {
		return nil, fmt.Errorf("创建 worker 池失败：%w", err)
	}

	return &ArgoCDGenerator{
		projectConfig: projectConfig,
		roleVars:      roleVars,
		outputDir:     outputDir,
		templateDir:   templateDir,
		workerPool:    pool,
	}, nil
}

// Close 关闭资源
func (g *ArgoCDGenerator) Close() error {
	if g.workerPool != nil {
		g.workerPool.Close()
	}
	return nil
}

// GenerateAll 生成所有 ArgoCD Application 配置
func (g *ArgoCDGenerator) GenerateAll() error {
	// 为每个应用生成 ArgoCD Application
	for _, roleVar := range g.roleVars {
		if err := g.GenerateForApp(roleVar.App); err != nil {
			return fmt.Errorf("生成 %s 失败：%w", roleVar.App, err)
		}
	}
	return nil
}

// GenerateForApp 为单个应用生成配置
func (g *ArgoCDGenerator) GenerateForApp(appName string) error {
	// 获取应用的 stack
	stackID, exists := g.projectConfig.Stack[appName]
	if !exists {
		return fmt.Errorf("应用 %s 未定义 Stack", appName)
	}

	// 构建渲染上下文
	ctx := map[string]interface{}{
		"project":      g.projectConfig.Project,
		"profile":      "int", // 或从配置读取
		"stack":        stackID,
		"item":         appName,
		"namespace":    "baas",
		"git_repo_url": g.buildGitRepoURL(appName),
		"git_branch":   "k8s_mas",
	}

	// 使用 Python Worker 渲染模板
	templatePath := filepath.Join(g.templateDir, "app.yaml.j2")
	content, err := g.workerPool.Render(template.RenderRequest{
		TemplatePath: templatePath,
		Context:      ctx,
	})
	if err != nil {
		return fmt.Errorf("渲染模板失败：%w", err)
	}

	// 写入文件
	// 文件名格式：{project}-{app}.yaml（与 Ansible 保持一致）
	// 示例：cms-cms-service.yaml
	outputFileName := fmt.Sprintf("%s-%s.yaml", g.projectConfig.Project, appName)
	outputPath := filepath.Join(g.outputDir, "argo-app", g.projectConfig.Project, "int", "k8s_"+stackID, outputFileName)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败：%w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入文件失败：%w", err)
	}

	fmt.Printf("✅ 已生成 ArgoCD Application: %s\n", outputPath)
	return nil
}

// buildGitRepoURL 构建 Git 仓库 URL
func (g *ArgoCDGenerator) buildGitRepoURL(appName string) string {
	// 从配置中构建完整的 Git 仓库 URL
	// 示例：https://github.example.com/{group}/{project}.git
	group := g.projectConfig.ToolsetGitGroup
	project := g.projectConfig.ToolsetGitProject

	if group == "" || project == "" {
		// 如果未配置，返回基础 URL
		return g.projectConfig.ToolsetGitBaseURL
	}

	return fmt.Sprintf("%s/%s/%s.git", g.projectConfig.ToolsetGitBaseURL, group, project)
}
