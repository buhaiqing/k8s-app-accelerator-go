package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/generator"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/model"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// generateCmd 生成命令
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "生成 K8s 配置",
	Long:  `根据配置文件和模板生成 Kubernetes 应用配置`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取 flags
		baseDir, _ := cmd.Flags().GetString("base-dir")
		workDir, _ := cmd.Flags().GetString("workdir")
		bootstrapName, _ := cmd.Flags().GetString("bootstrap")
		varsName, _ := cmd.Flags().GetString("vars")
		resourcesName, _ := cmd.Flags().GetString("resources")
		mappingName, _ := cmd.Flags().GetString("mapping")

		// 如果未指定工作目录，使用当前目录
		if workDir == "" {
			var err error
			workDir, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("获取当前目录失败：%w", err)
			}
		}

		// 转换工作目录为绝对路径
		absWorkDir, err := filepath.Abs(workDir)
		if err != nil {
			return fmt.Errorf("解析工作目录路径失败：%w", err)
		}
		workDir = absWorkDir

		fmt.Printf("================================================\n")
		fmt.Printf("工作目录：%s\n", workDir)
		fmt.Printf("基础目录：%s\n", baseDir)
		fmt.Printf("================================================\n")

		// 使用 base-dir 下的标准文件结构（支持自定义文件名）
		var configFile, resourcesFile, mappingFile, bootstrapFile string

		if varsName == "" {
			configFile = filepath.Join(baseDir, "vars.yaml")
		} else if !filepath.IsAbs(varsName) {
			configFile = filepath.Join(baseDir, varsName)
		} else {
			configFile = varsName
		}

		if resourcesName == "" {
			resourcesFile = filepath.Join(baseDir, "resources.yaml")
		} else if !filepath.IsAbs(resourcesName) {
			resourcesFile = filepath.Join(baseDir, resourcesName)
		} else {
			resourcesFile = resourcesName
		}

		if mappingName == "" {
			mappingFile = filepath.Join(baseDir, "mapping.yaml")
		} else if !filepath.IsAbs(mappingName) {
			mappingFile = filepath.Join(baseDir, mappingName)
		} else {
			mappingFile = mappingName
		}

		if bootstrapName == "" {
			bootstrapFile = filepath.Join(baseDir, "bootstrap.yml")
		} else if !filepath.IsAbs(bootstrapName) {
			bootstrapFile = filepath.Join(baseDir, bootstrapName)
		} else {
			bootstrapFile = bootstrapName
		}

		// 检查必需文件是否存在
		for _, file := range []string{configFile, resourcesFile, mappingFile} {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				return fmt.Errorf("文件不存在：%s", file)
			}
			fmt.Printf("✓ 找到配置文件：%s\n", filepath.Base(file))
		}

		// 加载配置
		loader := config.NewFileLoader()

		cfg, err := loader.LoadProjectConfig(configFile)
		if err != nil {
			return fmt.Errorf("加载配置文件失败：%w", err)
		}

		resources, err := loader.LoadResourceGroup(resourcesFile)
		if err != nil {
			return fmt.Errorf("加载资源文件失败：%w", err)
		}

		mapping, err := loader.LoadMapping(mappingFile)
		if err != nil {
			return fmt.Errorf("加载映射文件失败：%w", err)
		}

		// 加载 bootstrap.yml 获取 roleVars
		roleVars, err := loadBootstrap(bootstrapFile, workDir)
		if err != nil {
			// 如果 bootstrap.yml 不存在，使用默认 roleVars
			fmt.Printf("⚠ 未找到 bootstrap.yml，使用默认配置\n")
			roleVars = []*model.RoleVars{
				{
					App:       "cms-service",
					Type:      "backend",
					EnableHPA: true,
					EnableRDB: true,
					Resources: model.RoleResources{
						Default: model.ResourceConfig{
							LimitsCPU:      "1000m",
							LimitsMemory:   "1060Mi",
							RequestsCPU:    "10m",
							RequestsMemory: "960Mi",
						},
						Production: model.ResourceConfig{
							LimitsCPU:      "900m",
							LimitsMemory:   "3096Mi",
							RequestsCPU:    "300m",
							RequestsMemory: "3096Mi",
						},
					},
				},
			}
		}

		// 确定输出目录
		// 优先使用 vars.yaml 中的 rootdir，如果为空则使用工作目录下的 output
		outputDir := cfg.RootDir
		if outputDir == "" {
			outputDir = filepath.Join(workDir, "output")
		}
		// 如果是相对路径，转换为绝对路径（相对于工作目录）
		if !filepath.IsAbs(outputDir) {
			outputDir = filepath.Join(workDir, outputDir)
		}

		// 确定模板目录
		// 优先使用工作目录下的 templates，如果不存在则使用程序内置模板
		templateDir := filepath.Join(workDir, "templates")
		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			// 使用程序所在目录的 templates
			execPath, _ := os.Executable()
			templateDir = filepath.Join(filepath.Dir(execPath), "templates")
		}

		fmt.Printf("\n配置信息：\n")
		fmt.Printf("  - 项目：%s\n", cfg.Project)
		fmt.Printf("  - 环境：%v\n", cfg.Profiles)
		fmt.Printf("  - 输出目录：%s\n", outputDir)
		fmt.Printf("  - 模板目录：%s\n", templateDir)
		fmt.Printf("  - 应用数量：%d\n\n", len(roleVars))

		// 查找 Python Worker 脚本
		// 优先使用工作目录下的 scripts，如果不存在则使用程序所在目录的 scripts
		scriptPath := filepath.Join(workDir, "scripts", "render_worker.py")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			// 使用程序所在目录的 scripts
			execPath, _ := os.Executable()
			scriptPath = filepath.Join(filepath.Dir(execPath), "scripts", "render_worker.py")
			if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
				// 最后尝试当前工作目录
				scriptPath = "scripts/render_worker.py"
			}
		}
		fmt.Printf("  - Python Worker: %s\n\n", scriptPath)

		// 创建生成器
		gen, err := generator.NewGeneratorWithScript(cfg, resources, mapping, roleVars, outputDir, templateDir, scriptPath)
		if err != nil {
			return fmt.Errorf("创建生成器失败：%w", err)
		}
		defer gen.Close()

		// 生成配置
		if err := gen.GenerateAll(); err != nil {
			return fmt.Errorf("生成配置失败：%w", err)
		}

		fmt.Printf("\n================================================\n")
		fmt.Printf("✓ 配置生成完成！\n")
		fmt.Printf("输出目录：%s\n", outputDir)
		fmt.Printf("================================================\n")

		return nil
	},
}

func init() {
	// 添加 generate 命令到根命令
	rootCmd.AddCommand(generateCmd)

	// 添加 flags
	generateCmd.Flags().StringP("workdir", "w", "", "工作目录（所有配置文件的根目录，默认为当前目录）")
	generateCmd.Flags().String("bootstrap", "", "bootstrap 文件名（默认：bootstrap.yml）")
	generateCmd.Flags().String("vars", "", "vars 文件名（默认：configs/vars.yaml）")
	generateCmd.Flags().String("resources", "", "resources 文件名（默认：configs/resources.yaml）")
	generateCmd.Flags().String("mapping", "", "mapping 文件名（默认：mapping.yaml）")
}

// BootstrapFile bootstrap.yml 数据结构
type BootstrapFile struct {
	Roles []string `yaml:"roles"`
}

// loadBootstrap 加载 bootstrap.yml 文件和对应的 role vars
func loadBootstrap(bootstrapFile string, workDir string) ([]*model.RoleVars, error) {
	// 检查文件是否存在
	if _, err := os.Stat(bootstrapFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("bootstrap 文件不存在：%s", bootstrapFile)
	}

	// 读取文件
	data, err := os.ReadFile(bootstrapFile)
	if err != nil {
		return nil, fmt.Errorf("读取 bootstrap 文件失败：%w", err)
	}

	// 解析 YAML
	var bootstrap BootstrapFile
	if err := yaml.Unmarshal(data, &bootstrap); err != nil {
		return nil, fmt.Errorf("解析 bootstrap 文件失败：%w", err)
	}

	// 加载每个 role 的 vars
	roleVars := make([]*model.RoleVars, 0, len(bootstrap.Roles))
	for _, roleName := range bootstrap.Roles {
		// 尝试加载 roles/{roleName}/vars/main.yml
		roleVarsFile := filepath.Join(workDir, "roles", roleName, "vars", "main.yml")

		rv := &model.RoleVars{
			App: roleName,
		}

		if _, err := os.Stat(roleVarsFile); err == nil {
			// 加载 role vars
			varsData, err := os.ReadFile(roleVarsFile)
			if err == nil {
				yaml.Unmarshal(varsData, rv)
			}
		}

		// 设置默认值
		if rv.Type == "" {
			rv.Type = "backend"
		}
		// 默认启用 HPA 和 RDB（对于 backend 类型）
		if rv.Type == "backend" {
			if rv.EnableHPA == false {
				rv.EnableHPA = true
			}
			if rv.EnableRDB == false {
				rv.EnableRDB = true
			}
		}
		if rv.Resources.Default.LimitsCPU == "" {
			rv.Resources = model.RoleResources{
				Default: model.ResourceConfig{
					LimitsCPU:      "1000m",
					LimitsMemory:   "1060Mi",
					RequestsCPU:    "10m",
					RequestsMemory: "960Mi",
				},
				Production: model.ResourceConfig{
					LimitsCPU:      "900m",
					LimitsMemory:   "3096Mi",
					RequestsCPU:    "300m",
					RequestsMemory: "3096Mi",
				},
			}
		}

		roleVars = append(roleVars, rv)
	}

	return roleVars, nil
}
