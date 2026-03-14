package validator

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/model"
)

// ValidateProjectConfig 检查配置文件格式
func ValidateProjectConfig(cfg *config.ProjectConfig) *CheckReport {
	report := NewCheckReport()

	// 1. 项目名称不能为空
	if strings.TrimSpace(cfg.Project) == "" {
		report.AddError(
			"project",
			"项目名称不能为空",
			"在 vars.yaml 中配置 project 字段，例如：project: my-project",
		)
	} else {
		// 2. 项目名称格式（只能包含小写字母和数字）
		matched, _ := regexp.MatchString(`^[a-z0-9]+$`, cfg.Project)
		if !matched {
			report.AddWarning(
				"project",
				fmt.Sprintf("项目名称 '%s' 包含非小写字母或数字字符", cfg.Project),
				"建议使用小写字母和数字，例如：my-project-001",
			)
		}
	}

	// 3. 至少定义一个环境（profile）
	if len(cfg.Profiles) == 0 {
		report.AddError(
			"profiles",
			"至少需要定义一个环境（profile）",
			"在 vars.yaml 中添加 profiles 配置，例如：\nprofiles:\n  - int\n  - production",
		)
	} else {
		// 4. profile 名称规范性
		validProfiles := []string{"int", "uat", "production"}
		for _, profile := range cfg.Profiles {
			found := false
			for _, valid := range validProfiles {
				if profile == valid {
					found = true
					break
				}
			}
			if !found {
				report.AddWarning(
					fmt.Sprintf("profiles.%s", profile),
					fmt.Sprintf("环境名称 '%s' 不是标准环境名", profile),
					fmt.Sprintf("建议使用标准环境名：%s", strings.Join(validProfiles, ", ")),
				)
			}
		}
	}

	// 5. Apollo Token 格式验证
	if cfg.Apollo.Token != "" {
		if len(cfg.Apollo.Token) < 10 {
			report.AddWarning(
				"apollo.token",
				"Apollo Token 长度过短",
				"请确认 Apollo Token 是否正确复制",
			)
		}
	} else {
		report.AddWarning(
			"apollo.token",
			"Apollo Token 未配置",
			"如果需要使用 Apollo，请在 vars.yaml 中配置 apollo.token",
		)
	}

	return report
}

// ValidateResourceGroup 检查 Resources 完整性
func ValidateResourceGroup(resources *config.ResourceGroup) *CheckReport {
	report := NewCheckReport()

	// 1. 默认 RDS 连接地址必须配置
	hasDefaultRDS := false
	for _, rds := range resources.RDS {
		if rds.Name == "default" || rds.Name == "" {
			hasDefaultRDS = true
			if rds.DatasourceURL == "" {
				report.AddError(
					"rds.default.datasource_url",
					"默认 RDS 连接地址未配置",
					"在 resources.yaml 中配置 rds[0].datasource_url",
				)
			}

			// 2. 数据库端口范围
			if rds.DatasourcePort != "" {
				port, err := strconv.Atoi(rds.DatasourcePort)
				if err != nil || port < 1 || port > 65535 {
					report.AddError(
						"rds.default.datasource_port",
						fmt.Sprintf("数据库端口 '%s' 不合法", rds.DatasourcePort),
						"端口必须是 1-65535 之间的数字",
					)
				}
			}
		}
	}

	if !hasDefaultRDS && len(resources.RDS) > 0 {
		report.AddWarning(
			"rds",
			"未配置默认 RDS（name='default'）",
			"建议添加 name='default' 的 RDS 配置作为默认数据源",
		)
	}

	// 3. Redis 配置检查
	for _, redis := range resources.Redis {
		if redis.RedisIP == "" {
			report.AddWarning(
				"redis.redisIp",
				fmt.Sprintf("Redis '%s' 的 IP 地址未配置", redis.Name),
				"配置 redis.redisIp 字段",
			)
		}

		// 4. Redis 密码强度检查
		if redis.RedisPassword != "" && len(redis.RedisPassword) < 8 {
			report.AddWarning(
				"redis.redisPassword",
				"Redis 密码长度不足 8 位",
				"建议使用更强的密码（大小写 + 数字 + 特殊字符，长度≥12）",
			)
		}
	}

	// 5. OSS 配置检查
	for _, oss := range resources.OSS {
		if oss.BucketName == "" && oss.COSBucketName == "" {
			report.AddWarning(
				"oss.bucketName 或 cos.bucketName",
				fmt.Sprintf("OSS/COS '%s' 的 Bucket 名称未配置", oss.Name),
				"配置 bucketName 字段",
			)
		}
	}

	return report
}

// ValidateMappingConsistency 检查 Mapping 一致性
func ValidateMappingConsistency(apps []string, mapping *config.Mapping) *CheckReport {
	report := NewCheckReport()

	// 检查每个 app 是否在 mapping 中有定义
	for _, app := range apps {
		product, exists := mapping.Mappings[app]
		if !exists {
			report.AddError(
				fmt.Sprintf("mappings.%s", app),
				fmt.Sprintf("应用 '%s' 在 mapping.yaml 中没有定义", app),
				fmt.Sprintf("在 mapping.yaml 中添加映射关系：\nmappings:\n  %s: <product-name>", app),
			)
		} else if product == "" {
			report.AddError(
				fmt.Sprintf("mappings.%s", app),
				fmt.Sprintf("应用 '%s' 的 product 值为空", app),
				fmt.Sprintf("在 mapping.yaml 中为 %s 指定有效的 product 值", app),
			)
		} else {
			// 检查 product 格式
			matched, _ := regexp.MatchString(`^[a-z_]+$`, product)
			if !matched {
				report.AddWarning(
					fmt.Sprintf("mappings.%s", app),
					fmt.Sprintf("product '%s' 格式不规范", product),
					"product 应该只包含小写字母和下划线",
				)
			}
		}
	}

	return report
}

// ValidateRoleVars 检查 Role Vars 完整性
func ValidateRoleVars(roleVarsList []*model.RoleVars) *CheckReport {
	report := NewCheckReport()

	for i, roleVars := range roleVarsList {
		prefix := fmt.Sprintf("roles[%d]", i)

		// 1. app 字段必须定义
		if roleVars.App == "" {
			report.AddError(
				fmt.Sprintf("%s.app", prefix),
				"app 字段未定义",
				"在角色变量中配置 app 字段",
			)
		}

		// 2. DNET_PRODUCT 必须定义
		if roleVars.DNETProduct == "" {
			report.AddError(
				fmt.Sprintf("%s.DNET_PRODUCT", prefix),
				"DNET_PRODUCT 字段未定义",
				"在角色变量中配置 DNET_PRODUCT 字段",
			)
		}

		// 3. _type 只能是 backend 或 frontend
		if roleVars.Type != "" && roleVars.Type != "backend" && roleVars.Type != "frontend" {
			report.AddError(
				fmt.Sprintf("%s._type", prefix),
				fmt.Sprintf("_type 的值 '%s' 不合法", roleVars.Type),
				"_type 只能是 'backend' 或 'frontend'",
			)
		}

		// 4. 前端组件不应启用 enable_rdb
		if roleVars.Type == "frontend" && roleVars.EnableRDB {
			report.AddWarning(
				fmt.Sprintf("%s.enable_rdb", prefix),
				"前端组件不应该启用 enable_rdb",
				"将 enable_rdb 设置为 false 或删除该字段",
			)
		}

		// 5. CPU limits >= requests
		if roleVars.CPURequests != "" && roleVars.CPULimits != "" {
			if compareResource(roleVars.CPURequests, roleVars.CPULimits) > 0 {
				report.AddError(
					fmt.Sprintf("%s.cpu", prefix),
					fmt.Sprintf("CPU limits (%s) 小于 requests (%s)", roleVars.CPULimits, roleVars.CPURequests),
					"确保 CPU limits >= requests",
				)
			}
		}

		// 6. Memory limits >= requests
		if roleVars.MemoryRequests != "" && roleVars.MemoryLimits != "" {
			if compareResource(roleVars.MemoryRequests, roleVars.MemoryLimits) > 0 {
				report.AddError(
					fmt.Sprintf("%s.memory", prefix),
					fmt.Sprintf("Memory limits (%s) 小于 requests (%s)", roleVars.MemoryLimits, roleVars.MemoryRequests),
					"确保 Memory limits >= requests",
				)
			}
		}

		// 7. 内存请求合理性检查
		if roleVars.MemoryRequests != "" {
			memMB := parseMemoryToMB(roleVars.MemoryRequests)
			if memMB > 8192 {
				report.AddWarning(
					fmt.Sprintf("%s.memory_requests", prefix),
					fmt.Sprintf("内存请求过大 (%s)", roleVars.MemoryRequests),
					"确认是否真的需要这么多内存，考虑优化内存使用",
				)
			}
		}
	}

	return report
}

// ValidateTemplateFiles 检查模板文件存在性
func ValidateTemplateFiles(roleName string, profiles []string, templatesDir string) *CheckReport {
	report := NewCheckReport()

	// 必需的模板文件
	requiredTemplates := []string{
		"deployment.yaml.j2",
		"service.yaml.j2",
		"kustomization.yaml.j2",
		"config.yaml.j2",
	}

	// 检查每个 profile 的模板
	for _, profile := range profiles {
		templatePath := fmt.Sprintf("%s/%s/templates/overlays/%s", templatesDir, roleName, profile)

		for _, template := range requiredTemplates {
			fullPath := fmt.Sprintf("%s/%s", templatePath, template)
			if !fileExists(fullPath) {
				report.AddError(
					fmt.Sprintf("templates.%s.%s", profile, template),
					fmt.Sprintf("模板文件不存在：%s", fullPath),
					fmt.Sprintf("创建该模板文件或检查路径是否正确"),
				)
			}
		}

		// HPA 模板（条件必需）
		hpaPath := fmt.Sprintf("%s/hpa.yaml.j2", templatePath)
		// 这个检查需要知道 enable_hpa 的值，暂时跳过
		_ = hpaPath
	}

	return report
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// compareResource 比较资源大小（支持 CPU 和内存）
// 返回值：-1 表示 a < b, 0 表示 a == b, 1 表示 a > b
func compareResource(a, b string) int {
	// 简单实现，实际应该更复杂
	if a == b {
		return 0
	}

	// 尝试解析为数字
	aNum, errA := strconv.ParseFloat(strings.TrimRight(a, "m"), 64)
	bNum, errB := strconv.ParseFloat(strings.TrimRight(b, "m"), 64)

	if errA == nil && errB == nil {
		if aNum < bNum {
			return -1
		} else if aNum > bNum {
			return 1
		}
		return 0
	}

	// 字符串比较
	if a < b {
		return -1
	}
	return 1
}

// parseMemoryToMB 将内存字符串解析为 MB
func parseMemoryToMB(mem string) int {
	mem = strings.ToUpper(mem)

	if strings.HasSuffix(mem, "GI") {
		val, _ := strconv.Atoi(strings.TrimSuffix(mem, "GI"))
		return val * 1024
	} else if strings.HasSuffix(mem, "MI") {
		val, _ := strconv.Atoi(strings.TrimSuffix(mem, "MI"))
		return val
	} else if strings.HasSuffix(mem, "G") {
		val, _ := strconv.Atoi(strings.TrimSuffix(mem, "G"))
		return val * 1024
	} else if strings.HasSuffix(mem, "M") {
		val, _ := strconv.Atoi(strings.TrimSuffix(mem, "M"))
		return val
	}

	// 默认为字节
	val, _ := strconv.Atoi(mem)
	return val / 1024 / 1024
}

// CollectAllChecks 收集所有检查结果
func CollectAllChecks(
	cfg *config.ProjectConfig,
	resources *config.ResourceGroup,
	mapping *config.Mapping,
	roleVarsList []*model.RoleVars,
) *CheckReport {
	// 创建总报告
	totalReport := NewCheckReport()

	// A. 配置文件格式检查
	configReport := ValidateProjectConfig(cfg)
	mergeReport(totalReport, configReport)

	// B. Resources 完整性检查
	resourceReport := ValidateResourceGroup(resources)
	mergeReport(totalReport, resourceReport)

	// C. Mapping 一致性检查
	if roleVarsList != nil && len(roleVarsList) > 0 {
		apps := make([]string, len(roleVarsList))
		for i, rv := range roleVarsList {
			apps[i] = rv.App
		}
		mappingReport := ValidateMappingConsistency(apps, mapping)
		mergeReport(totalReport, mappingReport)
	}

	// D. Role Vars 完整性检查
	if roleVarsList != nil && len(roleVarsList) > 0 {
		roleVarsReport := ValidateRoleVars(roleVarsList)
		mergeReport(totalReport, roleVarsReport)
	}

	return totalReport
}

// mergeReport 合并多个报告
func mergeReport(target *CheckReport, source *CheckReport) {
	target.Results = append(target.Results, source.Results...)
	target.ErrorCount += source.ErrorCount
	target.WarnCount += source.WarnCount
	target.InfoCount += source.InfoCount
}
