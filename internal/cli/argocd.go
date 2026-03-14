package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/generator"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/model"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/validator"
	"github.com/spf13/cobra"
)

var (
	argocdOutputDir    string
	argocdRoles        []string
	skipArgoCDPrecheck bool
)

// argocdCmd 代表 argocd 命令
var argocdCmd = &cobra.Command{
	Use:   "argocd",
	Short: "生成 ArgoCD Application 配置",
	Long:  `生成 ArgoCD Application 配置，支持批量生成多个应用的 ArgoCD 配置`,
}

// generateArgoCDCmd 代表 generate-argocd 命令
var generateArgoCDCmd = &cobra.Command{
	Use:   "generate",
	Short: "生成 ArgoCD Application 配置",
	Long:  `根据配置文件和 bootstrap 文件生成 ArgoCD Application 配置`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerateArgoCD()
	},
}

func init() {
	// 添加 flags
	generateArgoCDCmd.Flags().StringVarP(&argocdOutputDir, "output", "o", "output", "输出目录")
	generateArgoCDCmd.Flags().StringSliceVar(&argocdRoles, "roles", nil, "指定要生成的 roles（逗号分隔）")
	generateArgoCDCmd.Flags().BoolVar(&skipArgoCDPrecheck, "skip-precheck", false, "跳过预检")

	// 添加全局 flags（从 generate.go 复制）
	// 注意：这里不需要 addCommonFlags，因为 flags 已经在 rootCmd 中定义

	// 将 generate-argocd 添加到 argocd 命令
	argocdCmd.AddCommand(generateArgoCDCmd)

	// 将 argocd 命令添加到 root
	rootCmd.AddCommand(argocdCmd)
}

func runGenerateArgoCD() error {
	verbose, _ := rootCmd.PersistentFlags().GetBool("verbose")
	if verbose {
		fmt.Println("🚀 开始生成 ArgoCD Application 配置...")
	}

	// 1. 加载配置
	configFile, _ := rootCmd.PersistentFlags().GetString("config")
	baseDir, _ := rootCmd.PersistentFlags().GetString("base-dir")

	// 处理相对路径
	if !filepath.IsAbs(configFile) {
		configFile = filepath.Join(baseDir, configFile)
	}

	projectConfig, err := func(configFile string) (*config.ProjectConfig, error) {
		loader := config.NewFileLoader()
		return loader.LoadProjectConfig(configFile)
	}(configFile)
	if err != nil {
		return fmt.Errorf("加载配置文件失败：%w", err)
	}

	// 2. 加载 bootstrap
	bootstrapFile, _ := rootCmd.PersistentFlags().GetString("bootstrap")
	if !filepath.IsAbs(bootstrapFile) {
		bootstrapFile = filepath.Join(baseDir, bootstrapFile)
	}

	roleVars, err := loadBootstrap(bootstrapFile, baseDir)
	if err != nil {
		return fmt.Errorf("加载 bootstrap 失败：%w", err)
	}

	// 如果指定了 roles，过滤只生成指定的 roles
	if len(argocdRoles) > 0 {
		filteredRoleVars := make([]*model.RoleVars, 0)
		for _, rv := range roleVars {
			for _, roleName := range argocdRoles {
				if rv.App == roleName {
					filteredRoleVars = append(filteredRoleVars, rv)
					break
				}
			}
		}
		if len(filteredRoleVars) == 0 {
			return fmt.Errorf("未找到指定的 roles: %v", argocdRoles)
		}
		roleVars = filteredRoleVars
	}

	// 3. Pre-Check（除非跳过）
	if !skipArgoCDPrecheck {
		fmt.Println("🔍 执行预检...")

		// 检查 ArgoCD 配置
		argocdResults := validator.ValidateArgoCDConfig(projectConfig)

		// 检查每个应用
		for _, rv := range roleVars {
			argocdResults = append(argocdResults, validator.ValidateArgoCDApplication(rv.App, projectConfig)...)
		}

		// 打印报告
		report := &validator.CheckReport{
			Results:    argocdResults,
			ErrorCount: 0,
			WarnCount:  0,
			InfoCount:  0,
		}
		for _, r := range argocdResults {
			if r.Level == "error" {
				report.ErrorCount++
			} else if r.Level == "warning" {
				report.WarnCount++
			} else if r.Level == "info" {
				report.InfoCount++
			}
		}
		report.PrintReport()

		// 如果有错误，退出
		if report.ErrorCount > 0 {
			return fmt.Errorf("预检发现 %d 个错误", report.ErrorCount)
		}
	}

	// 4. 创建输出目录
	outputPath := argocdOutputDir
	if !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(baseDir, outputPath)
	}

	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败：%w", err)
	}

	// 5. 创建生成器
	scriptPath := filepath.Join(baseDir, "scripts", "render_worker.py")
	templateDir := filepath.Join(baseDir, "templates", "argo-app")

	gen, err := generator.NewArgoCDGenerator(
		projectConfig,
		roleVars,
		outputPath,
		templateDir,
		scriptPath,
	)
	if err != nil {
		return fmt.Errorf("创建生成器失败：%w", err)
	}
	defer gen.Close()

	// 6. 生成配置
	fmt.Println("⚙️  正在生成 ArgoCD Application 配置...")
	if err := gen.GenerateAll(); err != nil {
		return fmt.Errorf("生成配置失败：%w", err)
	}

	fmt.Printf("✅ ArgoCD Application 配置生成完成！输出目录：%s\n", outputPath)
	return nil
}
