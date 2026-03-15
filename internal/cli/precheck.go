package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/validator"
	"github.com/spf13/cobra"
)

// precheckCmd 预检命令
var precheckCmd = &cobra.Command{
	Use:   "precheck",
	Short: "预检配置文件",
	Long:  `检查配置文件的完整性和正确性，提前发现潜在问题`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取 flags
		baseDir, _ := cmd.Flags().GetString("base-dir")
		configFile, _ := cmd.Flags().GetString("config")
		resourcesFile, _ := cmd.Flags().GetString("resources")
		mappingFile, _ := cmd.Flags().GetString("mapping")

		// 如果未指定 base-dir，使用默认值
		if baseDir == "" {
			baseDir = "/Users/bohaiqing/work/git/k8s_app_acelerator/gitlab_cfg"
		}

		// 如果配置文件未指定，使用 base-dir 下的默认文件
		if configFile == "" {
			configFile = filepath.Join(baseDir, "vars.yaml")
		} else if !filepath.IsAbs(configFile) {
			configFile = filepath.Join(baseDir, configFile)
		}

		if resourcesFile == "" {
			resourcesFile = filepath.Join(baseDir, "resources.yaml")
		} else if !filepath.IsAbs(resourcesFile) {
			resourcesFile = filepath.Join(baseDir, resourcesFile)
		}

		if mappingFile == "" {
			mappingFile = filepath.Join(baseDir, "mapping.yaml")
		} else if !filepath.IsAbs(mappingFile) {
			mappingFile = filepath.Join(baseDir, mappingFile)
		}

		// 检查文件是否存在
		for _, file := range []string{configFile, resourcesFile, mappingFile} {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				return fmt.Errorf("文件不存在：%s", file)
			}
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

		// 执行预检
		report := validator.CollectAllChecks(cfg, resources, mapping, nil)

		// 打印报告
		report.PrintReport()

		// 根据检查结果退出
		if !report.IsPassed() {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	// 添加 precheck 命令到根命令
	rootCmd.AddCommand(precheckCmd)

	// 添加 flags
	precheckCmd.Flags().StringP("config", "c", "", "配置文件路径 (vars.yaml)，支持相对于 base-dir 的路径")
	precheckCmd.Flags().String("resources", "", "资源文件路径 (resources.yaml)，支持相对于 base-dir 的路径")
	precheckCmd.Flags().String("mapping", "", "映射文件路径 (mapping.yaml)，支持相对于 base-dir 的路径")
}
