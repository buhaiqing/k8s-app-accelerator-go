package template_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/template"
)

// TestWorkerPool_Integration 集成测试
func TestWorkerPool_Integration(t *testing.T) {
	// 创建临时模板文件
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.yaml.j2")

	templateContent := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ app_name }}
  namespace: {{ namespace | default("default") }}
data:
  key: {{ value }}
`
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	assert.NoError(t, err)

	// 创建 Worker 池（只有 1 个 worker 方便调试）
	// 使用 ../../scripts/ 因为测试在 internal/template/ 目录下
	pool, err := template.NewWorkerPool(1, "../../scripts/render_worker.py")
	assert.NoError(t, err)
	defer pool.Close()

	// 检查健康状态
	assert.Equal(t, 1, pool.Size())
	assert.Equal(t, 1, pool.HealthCheck())

	// 渲染模板
	content, err := pool.Render(template.RenderRequest{
		TemplatePath: templatePath,
		Context: map[string]interface{}{
			"app_name":  "test-app",
			"namespace": "production",
			"value":     "test-value",
		},
	})

	assert.NoError(t, err)
	assert.Contains(t, content, "name: test-app")
	assert.Contains(t, content, "namespace: production")
	assert.Contains(t, content, "key: test-value")
}
