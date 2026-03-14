package generator_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/generator"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/model"
)

// TestNewGenerator_Basic 测试基本创建（跳过实际 Worker 启动）
func TestNewGenerator_Basic(t *testing.T) {
	t.Skip("跳过需要实际 Python 环境的测试")

	cfg := &config.ProjectConfig{
		Project:  "test-project",
		Profiles: []string{"int"},
	}

	resources := &config.ResourceGroup{}
	mapping := &config.Mapping{}
	roleVars := []*model.RoleVars{}

	gen, err := generator.NewGenerator(cfg, resources, mapping, roleVars, "output", "../../scripts/render_worker.py")

	assert.NoError(t, err)
	assert.NotNil(t, gen)

	defer gen.Close()
}

// TestGenerator_Structure 测试结构体字段
func TestGenerator_Structure(t *testing.T) {
	// 创建一个不启动 Worker 的生成器（用于测试结构）
	gen := &generator.Generator{}

	assert.NotNil(t, gen)
}

// TestDataStructures 测试数据结构
func TestDataStructures(t *testing.T) {
	// 测试 ProjectConfig
	cfg := &config.ProjectConfig{
		Project:  "test-project",
		Profiles: []string{"int", "production"},
	}

	assert.Equal(t, "test-project", cfg.Project)
	assert.Len(t, cfg.Profiles, 2)

	// 测试 RoleVars
	roleVars := &model.RoleVars{
		App:            "test-app",
		DNETProduct:    "test-product",
		Type:           "backend",
		CPURequests:    "100m",
		MemoryRequests: "256M",
	}

	assert.Equal(t, "test-app", roleVars.App)
	assert.Equal(t, "backend", roleVars.Type)
}
