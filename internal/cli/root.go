package cli

import (
	"github.com/spf13/cobra"
)

// rootCmd 是根命令
var rootCmd = &cobra.Command{
	Use:   "k8s-gen",
	Short: "K8s 应用配置生成器",
	Long: `基于 Ansible roles 模板生成 Kubernetes 应用配置
保持 100% Jinja2 模板兼容性，同时获得 Go语言的性能优势`,
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 添加全局 flags
	rootCmd.PersistentFlags().StringP("base-dir", "b", "/Users/bohaiqing/work/git/k8s_app_acelerator/gitlab_cfg", "基础目录路径（默认会读取该目录下的 bootstrap.yml, configs/vars.yaml, configs/resources.yaml, configs/mapping.yaml）")
}
