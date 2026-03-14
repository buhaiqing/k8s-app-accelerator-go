package validator

import (
	"fmt"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
)

// ValidateArgoCDConfig 验证 ArgoCD 配置
func ValidateArgoCDConfig(projectConfig *config.ProjectConfig) []CheckResult {
	var results []CheckResult

	// A. ArgoCD Site 检查
	if projectConfig.ArgoCD.Site == "" {
		results = append(results, CheckResult{
			Level:      "error",
			Field:      "argocd.site",
			Message:    "ArgoCD Site 不能为空",
			Suggestion: "配置 ArgoCD 站点地址，如：https://argocd.example.com",
		})
	}

	// B. Git 仓库 URL 检查
	if projectConfig.ToolsetGitBaseURL == "" {
		results = append(results, CheckResult{
			Level:      "error",
			Field:      "toolset_git_base_url",
			Message:    "Git 仓库 URL 不能为空",
			Suggestion: "配置 Git 仓库地址，如：https://github.example.com/org/repo.git",
		})
	}

	// C. Stack 映射检查
	if len(projectConfig.Stack) == 0 {
		results = append(results, CheckResult{
			Level:      "warning",
			Field:      "stack",
			Message:    "未定义任何 Stack 映射",
			Suggestion: "至少配置一个应用的 Stack 映射",
		})
	}

	return results
}

// ValidateArgoCDApplication 验证 ArgoCD Application 生成
func ValidateArgoCDApplication(appName string, projectConfig *config.ProjectConfig) []CheckResult {
	var results []CheckResult

	// 检查 Stack 是否定义
	if _, exists := projectConfig.Stack[appName]; !exists {
		results = append(results, CheckResult{
			Level:      "error",
			Field:      fmt.Sprintf("stack.%s", appName),
			Message:    fmt.Sprintf("应用 %s 未定义 Stack", appName),
			Suggestion: fmt.Sprintf("在 stack 配置中添加 %s 的映射", appName),
		})
	}

	return results
}
