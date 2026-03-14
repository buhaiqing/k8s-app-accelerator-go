package validator_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/model"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/validator"
)

// TestValidateProjectConfig_Basic 测试基本配置检查
func TestValidateProjectConfig_Basic(t *testing.T) {
	// 测试空配置
	cfg := &config.ProjectConfig{
		Project:  "",
		Profiles: []string{},
	}

	report := validator.ValidateProjectConfig(cfg)

	assert.False(t, report.IsPassed())
	assert.Greater(t, report.ErrorCount, 0)
}

// TestValidateProjectConfig_Valid 测试有效配置
func TestValidateProjectConfig_Valid(t *testing.T) {
	cfg := &config.ProjectConfig{
		Project:  "test-project",
		Profiles: []string{"int", "production"},
		Apollo: config.ApolloConfig{
			Token: "abcdefghijklmnopqrstuvwxyz123456",
		},
	}

	report := validator.ValidateProjectConfig(cfg)

	assert.True(t, report.IsPassed())
	assert.Equal(t, 0, report.ErrorCount)
}

// TestValidateResourceGroup_RDS 测试 RDS 资源检查
func TestValidateResourceGroup_RDS(t *testing.T) {
	resources := &config.ResourceGroup{
		RDS: []config.RDSResource{
			{
				Name:           "default",
				DatasourceURL:  "",      // 错误：空值
				DatasourcePort: "99999", // 错误：端口超出范围
			},
		},
	}

	report := validator.ValidateResourceGroup(resources)

	assert.False(t, report.IsPassed())
	assert.Greater(t, report.ErrorCount, 0)
}

// TestValidateMappingConsistency 测试 Mapping 一致性
func TestValidateMappingConsistency(t *testing.T) {
	apps := []string{"cms-service", "fms-service"}
	mapping := &config.Mapping{
		Mappings: map[string]string{
			"cms-service": "cms",
			// fms-service 缺失
		},
	}

	report := validator.ValidateMappingConsistency(apps, mapping)

	assert.False(t, report.IsPassed())
	assert.Greater(t, report.ErrorCount, 0)
	assert.Contains(t, report.Results[0].Message, "在 mapping.yaml 中没有定义")
}

// TestValidateRoleVars_Complete 测试完整的 Role Vars
func TestValidateRoleVars_Complete(t *testing.T) {
	roleVarsList := []*model.RoleVars{
		{
			App:            "test-app",
			DNETProduct:    "test-product",
			Type:           "backend",
			EnableHPA:      true,
			EnableRDB:      false,
			CPURequests:    "100m",
			CPULimits:      "200m",
			MemoryRequests: "256M",
			MemoryLimits:   "512M",
		},
	}

	report := validator.ValidateRoleVars(roleVarsList)

	assert.True(t, report.IsPassed())
	assert.Equal(t, 0, report.ErrorCount)
}

// TestValidateRoleVars_Invalid 测试无效的 Role Vars
func TestValidateRoleVars_Invalid(t *testing.T) {
	roleVarsList := []*model.RoleVars{
		{
			App:         "",        // 错误：空值
			DNETProduct: "",        // 错误：空值
			Type:        "invalid", // 错误：无效类型
			CPURequests: "500m",
			CPULimits:   "200m", // 错误：limits < requests
		},
	}

	report := validator.ValidateRoleVars(roleVarsList)

	assert.False(t, report.IsPassed())
	assert.Greater(t, report.ErrorCount, 0)
}

// TestCollectAllChecks_Integration 集成测试
func TestCollectAllChecks_Integration(t *testing.T) {
	cfg := &config.ProjectConfig{
		Project:  "test-project",
		Profiles: []string{"int"},
	}

	resources := &config.ResourceGroup{
		RDS: []config.RDSResource{
			{
				Name:          "default",
				DatasourceURL: "rm-test.mysql.rds.aliyuncs.com",
			},
		},
	}

	mapping := &config.Mapping{
		Mappings: map[string]string{
			"test-app": "test",
		},
	}

	roleVarsList := []*model.RoleVars{
		{
			App:         "test-app",
			DNETProduct: "test-product",
		},
	}

	report := validator.CollectAllChecks(cfg, resources, mapping, roleVarsList)

	assert.NotNil(t, report)
	assert.True(t, report.IsPassed())
}
