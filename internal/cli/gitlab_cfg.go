package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/generator"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	gitlabCfgOutputDir    string
	gitlabCfgRoles        []string
	skipGitlabCfgPrecheck bool
)

// gitlabCfgCmd 代表 gitlab-cfg 命令
var gitlabCfgCmd = &cobra.Command{
	Use:   "gitlab-cfg",
	Short: "生成 GitLab 项目配置",
	Long:  `生成 GitLab 项目配置，支持批量生成多个项目的 GitLab CI/CD 配置`,
}

// generateGitlabCfgCmd 代表 generate 命令
var generateGitlabCfgCmd = &cobra.Command{
	Use:   "generate",
	Short: "生成 GitLab 项目配置",
	Long:  `根据配置文件和模板生成 GitLab 项目配置`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerateGitlabCfg()
	},
}

func init() {
	// 添加 flags
	generateGitlabCfgCmd.Flags().StringVarP(&gitlabCfgOutputDir, "output", "o", "output", "输出目录")
	generateGitlabCfgCmd.Flags().StringSliceVar(&gitlabCfgRoles, "roles", nil, "指定要生成的 roles（逗号分隔）")
	generateGitlabCfgCmd.Flags().BoolVar(&skipGitlabCfgPrecheck, "skip-precheck", false, "跳过预检")

	// 将 generate 添加到 gitlab-cfg 命令
	gitlabCfgCmd.AddCommand(generateGitlabCfgCmd)

	// 将 gitlab-cfg 命令添加到 root
	rootCmd.AddCommand(gitlabCfgCmd)
}

func runGenerateGitlabCfg() error {
	verbose, _ := rootCmd.PersistentFlags().GetBool("verbose")
	if verbose {
		fmt.Println("🚀 开始生成 GitLab 项目配置...")
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

	// 2. Pre-Check（除非跳过）
	if !skipGitlabCfgPrecheck {
		fmt.Println("🔍 执行预检...")

		// 验证关键字段
		if err := validateProjectConfig(projectConfig); err != nil {
			return err
		}

		fmt.Println("✅ 预检通过")
	}

	// 3. 创建输出目录
	// outputPath 使用 Go 项目中的 output 目录（相对于当前工作目录）
	outputPath := filepath.Join(".", gitlabCfgOutputDir)

	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败：%w", err)
	}

	// 4. 创建生成器
	// scriptPath：使用 Go 项目中的 Python 渲染脚本
	scriptPath := filepath.Join(".", "scripts", "render_worker.py")
	// templateDir：直接使用 Ansible roles 目录，不管理模板
	templateDir := filepath.Join(baseDir, "roles")

	// 加载 resources 和 mapping
	resourcesFile, _ := rootCmd.PersistentFlags().GetString("resources")
	mappingFile, _ := rootCmd.PersistentFlags().GetString("mapping")

	if !filepath.IsAbs(resourcesFile) {
		resourcesFile = filepath.Join(baseDir, resourcesFile)
	}
	if !filepath.IsAbs(mappingFile) {
		mappingFile = filepath.Join(baseDir, mappingFile)
	}

	loader := config.NewFileLoader()

	// 加载 resources（使用 map[string]interface{}接收）
	var resources map[string]interface{}
	resourcesData, err := loader.LoadResourceGroup(resourcesFile)
	if err != nil {
		return fmt.Errorf("加载 resources 文件失败：%w", err)
	}
	// 将 ResourceGroup 转换为 map[string]interface{}
	resourcesBytes, _ := yaml.Marshal(resourcesData)
	yaml.Unmarshal(resourcesBytes, &resources)

	// 加载 mapping
	mappingData, err := loader.LoadMapping(mappingFile)
	if err != nil {
		return fmt.Errorf("加载 mapping 文件失败：%w", err)
	}
	// 将 Mapping 转换为 map[string]string
	mapping := mappingData.Mappings

	gen, err := generator.NewGitlabCfgGenerator(
		projectConfig,
		outputPath,
		templateDir,
		scriptPath,
		resources,
		mapping,
	)
	if err != nil {
		return fmt.Errorf("创建生成器失败：%w", err)
	}

	// 5. 生成配置
	fmt.Println("📝 开始生成配置...")
	if err := gen.GenerateAll(); err != nil {
		return fmt.Errorf("生成配置失败：%w", err)
	}

	if verbose {
		fmt.Printf("✅ GitLab 配置生成完成，输出目录：%s\n", outputPath)
	}

	return nil
}

// validateProjectConfig 验证项目配置的关键字段
func validateProjectConfig(cfg *config.ProjectConfig) error {
	var errors []string

	// 验证 DNET_PRODUCT
	if cfg.DNETProduct == "" {
		errors = append(errors, "❌ DNET_PRODUCT 未定义或为空\n   解决方案：请在配置文件中设置 'product' 字段，例如：product: cms")
	}

	// 验证 Profiles
	if len(cfg.Profiles) == 0 {
		errors = append(errors, "❌ profiles 列表为空\n   解决方案：请在配置文件中设置 'profiles' 字段，例如：profiles:\n     - production\n     - int")
	} else {
		// 检查 profiles 中是否有空值
		for i, profile := range cfg.Profiles {
			if profile == "" {
				errors = append(errors, fmt.Sprintf("❌ profiles[%d] 为空字符串\n   解决方案：请移除空字符串或填充有效的 profile 名称", i))
			}
		}
	}

	// 验证 Stack
	if len(cfg.Stack) == 0 {
		errors = append(errors, "❌ stack 字段为空\n   解决方案：请在配置文件中设置 'stack' 字段，例如：stack:\n     cms-service: baas")
	} else {
		// 检查每个应用的 stack 值
		for appName, stackID := range cfg.Stack {
			if stackID == "" {
				errors = append(errors, fmt.Sprintf("❌ 应用 '%s' 的 stack 值为空\n   解决方案：请为应用 '%s' 指定有效的 stack ID", appName, appName))
			}
		}
	}

	// 如果有错误，返回合并的错误信息
	if len(errors) > 0 {
		errMsg := "\n\n========================================\n" +
			"   🚨 配置验证失败\n" +
			"========================================\n\n"
		for _, e := range errors {
			errMsg += e + "\n\n"
		}
		errMsg += "========================================\n" +
			"提示：请检查配置文件 (vars.yaml) 并修正上述问题\n" +
			"========================================\n"
		return fmt.Errorf("%s", errMsg)
	}

	// 警告信息（不阻止执行）
	if cfg.Project == "" {
		fmt.Printf("⚠️  警告：project 未设置，将使用默认值 'default'\n")
	}

	return nil
}
