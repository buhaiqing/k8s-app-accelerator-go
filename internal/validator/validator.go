package validator

import (
	"github.com/fatih/color"
)

// CheckLevel 检查级别
type CheckLevel string

const (
	LevelError   CheckLevel = "error"
	LevelWarning CheckLevel = "warning"
	LevelInfo    CheckLevel = "info"
)

// CheckResult 单个检查结果
type CheckResult struct {
	Level      CheckLevel `json:"level"`
	Field      string     `json:"field"`
	Message    string     `json:"message"`
	Suggestion string     `json:"suggestion,omitempty"`
}

// CheckReport 检查报告
type CheckReport struct {
	Results    []CheckResult `json:"results"`
	ErrorCount int           `json:"error_count"`
	WarnCount  int           `json:"warn_count"`
	InfoCount  int           `json:"info_count"`
	Passed     bool          `json:"passed"`
}

// NewCheckReport 创建新的检查报告
func NewCheckReport() *CheckReport {
	return &CheckReport{
		Results: make([]CheckResult, 0),
	}
}

// AddError 添加错误级别的检查结果
func (r *CheckReport) AddError(field, message, suggestion string) {
	r.Results = append(r.Results, CheckResult{
		Level:      LevelError,
		Field:      field,
		Message:    message,
		Suggestion: suggestion,
	})
	r.ErrorCount++
}

// AddWarning 添加警告级别的检查结果
func (r *CheckReport) AddWarning(field, message, suggestion string) {
	r.Results = append(r.Results, CheckResult{
		Level:      LevelWarning,
		Field:      field,
		Message:    message,
		Suggestion: suggestion,
	})
	r.WarnCount++
}

// AddInfo 添加信息级别的检查结果
func (r *CheckReport) AddInfo(field, message, suggestion string) {
	r.Results = append(r.Results, CheckResult{
		Level:      LevelInfo,
		Field:      field,
		Message:    message,
		Suggestion: suggestion,
	})
	r.InfoCount++
}

// IsPassed 检查是否通过（没有错误）
func (r *CheckReport) IsPassed() bool {
	return r.ErrorCount == 0
}

// PrintReport 打印彩色报告
func (r *CheckReport) PrintReport() {
	if r.ErrorCount > 0 {
		color.Red("❌ 发现 %d 个错误", r.ErrorCount)
		for _, result := range r.Results {
			if result.Level == LevelError {
				color.Red("  ✖ %s\n", result.Field)
				color.Red("    问题：%s\n", result.Message)
				if result.Suggestion != "" {
					color.Red("    建议：%s\n\n", result.Suggestion)
				}
			}
		}
	}

	if r.WarnCount > 0 {
		color.Yellow("\n⚠️  发现 %d 个警告", r.WarnCount)
		for _, result := range r.Results {
			if result.Level == LevelWarning {
				color.Yellow("  ⚠ %s\n", result.Field)
				color.Yellow("    问题：%s\n", result.Message)
				if result.Suggestion != "" {
					color.Yellow("    建议：%s\n\n", result.Suggestion)
				}
			}
		}
	}

	if r.InfoCount > 0 {
		color.Cyan("\nℹ️  %d 条提示", r.InfoCount)
		for _, result := range r.Results {
			if result.Level == LevelInfo {
				color.Cyan("  ℹ %s\n", result.Field)
				color.Cyan("    %s\n\n", result.Message)
			}
		}
	}

	// 总结
	if r.IsPassed() {
		color.Green("\n✓ 预检通过\n")
	} else {
		color.Red("\n✗ 预检未通过，请修复上述错误\n")
	}
}
