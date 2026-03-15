package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/generator"
	"github.com/spf13/cobra"
)

var (
	jenkinsOutputDir    string
	jenkinsRoles        []string
	skipJenkinsPrecheck bool
)

// jenkinsCmd 代表 jenkins 命令
var jenkinsCmd = &cobra.Command{
	Use:   "jenkins",
	Short: "生成 Jenkins Jobs 配置",
	Long:  `生成 Jenkins Jobs 配置，支持批量生成多个产品的 Jenkins Job 配置`,
}

// generateJenkinsCmd 代表 generate-jenkins 命令
var generateJenkinsCmd = &cobra.Command{
	Use:   "generate",
	Short: "生成 Jenkins Jobs 配置",
	Long:  `根据配置文件和模板生成 Jenkins Jobs 配置`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerateJenkins()
	},
}

func init() {
	// 添加 flags
	generateJenkinsCmd.Flags().StringVarP(&jenkinsOutputDir, "output", "o", "output", "输出目录")
	generateJenkinsCmd.Flags().StringSliceVar(&jenkinsRoles, "roles", nil, "指定要生成的 roles（逗号分隔）")
	generateJenkinsCmd.Flags().BoolVar(&skipJenkinsPrecheck, "skip-precheck", false, "跳过预检")

	// 将 generate-jenkins 添加到 jenkins 命令
	jenkinsCmd.AddCommand(generateJenkinsCmd)

	// 将 jenkins 命令添加到 root
	rootCmd.AddCommand(jenkinsCmd)
}

func runGenerateJenkins() error {
	verbose, _ := rootCmd.PersistentFlags().GetBool("verbose")
	if verbose {
		fmt.Println("🚀 开始生成 Jenkins Jobs 配置...")
	}

	// 1. 加载配置（复用 Ansible 的 vars.yaml）
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

	// 2. Pre-Check（除非跳过）
	if !skipJenkinsPrecheck {
		fmt.Println("🔍 执行预检...")
		// TODO: 实现 Jenkins 预检逻辑
		fmt.Println("⚠️  Jenkins 预检功能尚未实现")
	}

	// 3. 创建输出目录
	// outputPath 使用绝对路径
	outputPath := jenkinsOutputDir
	if !filepath.IsAbs(outputPath) {
		// 转换为绝对路径（相对于当前工作目录）
		cwd, _ := os.Getwd()
		outputPath = filepath.Join(cwd, jenkinsOutputDir)
	}

	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败：%w", err)
	}

	// 4. 创建生成器
	// scriptPath：使用 Go 项目中的 Python 渲染脚本
	scriptPath := filepath.Join(".", "scripts", "render_worker.py")
	// templateDir：直接使用 Ansible roles 目录，不管理模板
	templateDir := filepath.Join(baseDir, "roles")

	gen, err := generator.NewJenkinsGenerator(
		projectConfig,
		outputPath,
		templateDir,
		scriptPath,
	)
	if err != nil {
		return fmt.Errorf("创建生成器失败：%w", err)
	}
	defer gen.Close()

	// 5. 生成配置
	fmt.Println("⚙️  正在生成 Jenkins Jobs 配置...")
	if err := gen.GenerateAll(); err != nil {
		return fmt.Errorf("生成配置失败：%w", err)
	}

	fmt.Printf("✅ Jenkins Jobs 配置生成完成！输出目录：%s\n", outputPath)
	return nil
}
