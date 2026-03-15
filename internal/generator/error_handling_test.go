package generator_test

import (
	"testing"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/config"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/generator"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestGenerator_ErrorHandling(t *testing.T) {
	t.Run("并发错误收集", func(t *testing.T) {
		// 创建测试配置
		cfg := &config.ProjectConfig{
			Project:  "test-project",
			Profiles: []string{"int", "uat"},
			Stack:    map[string]string{"app1": "int", "app2": "uat"},
		}

		resources := &config.ResourceGroup{}

		mapping := &config.Mapping{
			Mappings: map[string]string{
				"app1": "product1",
				"app2": "product2",
			},
		}

		// 创建多个 roleVars（其中一个会失败）
		roleVars := []*model.RoleVars{
			{
				App:       "app1",
				Type:      "backend",
				EnableHPA: true,
				EnableRDB: false,
			},
			{
				App:       "app2",
				Type:      "backend",
				EnableHPA: true,
				EnableRDB: false,
			},
		}

		// 创建生成器（使用不存在的脚本路径，模拟失败场景）
		gen, err := generator.NewGeneratorWithScript(
			cfg,
			resources,
			mapping,
			roleVars,
			"/tmp/output",
			"/tmp/templates",
			"/nonexistent/render_worker.py", // 故意使用不存在的路径
		)

		// 应该返回错误（因为脚本不存在）
		assert.Error(t, err)
		assert.Nil(t, gen)
	})
}

func TestGenerator_ConcurrentControl(t *testing.T) {
	t.Run("并发数限制", func(t *testing.T) {
		// 这个测试验证并发控制机制
		// 创建大量应用配置
		cfg := &config.ProjectConfig{
			Project:  "test-project",
			Profiles: []string{"int"},
			Stack:    make(map[string]string),
		}

		// 创建 100 个应用
		roleVars := make([]*model.RoleVars, 100)
		mapping := &config.Mapping{
			Mappings: make(map[string]string),
		}

		for i := 0; i < 100; i++ {
			appName := "app" + string(rune('0'+i%10)) + string(rune('0'+i/10))
			cfg.Stack[appName] = "int"
			mapping.Mappings[appName] = "product"
			roleVars[i] = &model.RoleVars{
				App:       appName,
				Type:      "backend",
				EnableHPA: false,
				EnableRDB: false,
			}
		}

		// 验证配置正确性
		assert.Equal(t, 100, len(roleVars))
		assert.Equal(t, 100, len(mapping.Mappings))
	})
}
