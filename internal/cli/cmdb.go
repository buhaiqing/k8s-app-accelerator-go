package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/generator"
	"github.com/spf13/cobra"
)

// cmdbCmd cmdb 命令
var cmdbCmd = &cobra.Command{
	Use:   "cmdb",
	Short: "生成 CMDB 初始化 SQL",
	Long:  `根据配置文件生成 CMDB 数据库初始化 SQL 脚本`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取 flags
		baseDir, _ := cmd.Flags().GetString("base-dir")
		workDir, _ := cmd.Flags().GetString("workdir")
		varsName, _ := cmd.Flags().GetString("vars")
		resourcesName, _ := cmd.Flags().GetString("resources")
		outputOverride, _ := cmd.Flags().GetString("output")

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

		// 确定配置文件路径
		var configFile, resourcesFile string

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

		// 检查必需文件是否存在
		for _, file := range []string{configFile, resourcesFile} {
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

		// 确定输出目录
		outputDir := cfg.RootDir
		if outputOverride != "" {
			outputDir = outputOverride
		}
		if outputDir == "" {
			outputDir = filepath.Join(workDir, "out")
		}
		// 如果是相对路径，转换为绝对路径（相对于工作目录）
		if !filepath.IsAbs(outputDir) {
			outputDir = filepath.Join(workDir, outputDir)
		}

		// 确定模板目录
		// 优先使用工作目录下的 templates，如果不存在则使用程序所在目录的 templates
		templateDir := filepath.Join(workDir, "templates")
		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			// 使用程序所在目录的 templates
			execPath, _ := os.Executable()
			templateDir = filepath.Join(filepath.Dir(execPath), "templates")
		}
		// 最后尝试当前目录的 templates
		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			templateDir = "templates"
		}

		fmt.Printf("\n配置信息：\n")
		fmt.Printf("  - 项目：%s\n", cfg.Project)
		fmt.Printf("  - 环境：%v\n", cfg.Profiles)
		fmt.Printf("  - 输出目录：%s\n", outputDir)
		fmt.Printf("  - 模板目录：%s\n", templateDir)

		// 查找 Python Worker 脚本
		scriptPath := filepath.Join(workDir, "scripts", "render_worker.py")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			execPath, _ := os.Executable()
			scriptPath = filepath.Join(filepath.Dir(execPath), "scripts", "render_worker.py")
			if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
				scriptPath = "scripts/render_worker.py"
			}
		}
		fmt.Printf("  - Python Worker: %s\n\n", scriptPath)

		// 创建 CMDB 生成器
		gen, err := generator.NewCMDBGenerator(cfg, resources, outputDir, templateDir, scriptPath)
		if err != nil {
			return fmt.Errorf("创建 CMDB 生成器失败：%w", err)
		}
		defer gen.Close()

		// 生成 SQL 配置
		if err := gen.GenerateAll(); err != nil {
			return fmt.Errorf("生成 SQL 配置失败：%w", err)
		}

		fmt.Printf("\n================================================\n")
		fmt.Printf("✓ SQL 配置生成完成！\n")
		fmt.Printf("输出目录：%s\n", outputDir)
		fmt.Printf("================================================\n")

		return nil
	},
}

func init() {
	// 添加 cmdb 命令到根命令
	rootCmd.AddCommand(cmdbCmd)

	// 添加 flags
	cmdbCmd.Flags().StringP("workdir", "w", "", "工作目录（所有配置文件的根目录，默认为当前目录）")
	cmdbCmd.Flags().String("vars", "", "vars 文件名（默认：configs/vars.yaml）")
	cmdbCmd.Flags().String("resources", "", "resources 文件名（默认：configs/resources.yaml）")
	cmdbCmd.Flags().StringP("output", "o", "", "输出目录（默认使用 vars.yaml 中的 rootdir 或 ./out）")
}
