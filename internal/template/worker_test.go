package template_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/buhaiqing/k8s-app-accelerator-go/internal/template"
)

// TestRenderWorker_SingleRender 测试单次渲染
func TestRenderWorker_SingleRender(t *testing.T) {
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

	// 创建 Worker
	// 使用 ../../scripts/ 因为测试在 internal/template/ 目录下
	worker, err := template.NewPythonWorker("../../scripts/render_worker.py")
	assert.NoError(t, err)
	defer worker.Close()

	// 渲染模板
	content, err := worker.Render(template.RenderRequest{
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

// TestRenderWorker_WithFilters 测试 Filters
func TestRenderWorker_WithFilters(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "filter_test.yaml.j2")

	templateContent := `
mandatory: {{ mandatory_value | mandatory }}
profile: {{ profile | profile_convert }}
ternary: {{ enable | ternary('enabled', 'disabled') }}
upper: {{ text | upper }}
lower: {{ text | lower }}
first: {{ items | first }}
last: {{ items | last }}
count: {{ items | count }}
unique: {{ duplicates | unique | join(',') }}
combine: {{ (dict1 | combine(dict2)).key1 }},{{ (dict1 | combine(dict2)).key3 }}
`
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	assert.NoError(t, err)

	worker, err := template.NewPythonWorker("../../scripts/render_worker.py")
	assert.NoError(t, err)
	defer worker.Close()

	content, err := worker.Render(template.RenderRequest{
		TemplatePath: templatePath,
		Context: map[string]interface{}{
			"mandatory_value": "required",
			"profile":         "int",
			"enable":          true,
			"text":            "Hello",
			"items":           []string{"a", "b", "c"},
			"duplicates":      []string{"a", "b", "a", "c", "b"},
			"dict1":           map[string]string{"key1": "value1", "key2": "value2"},
			"dict2":           map[string]string{"key2": "value2-updated", "key3": "value3"},
		},
	})

	assert.NoError(t, err)
	assert.Contains(t, content, "mandatory: required")
	assert.Contains(t, content, "profile: Int") // 与 Ansible 兼容：int -> Int
	assert.Contains(t, content, "ternary: enabled")
	assert.Contains(t, content, "upper: HELLO")
	assert.Contains(t, content, "lower: hello")
	assert.Contains(t, content, "first: a")
	assert.Contains(t, content, "last: c")
	assert.Contains(t, content, "count: 3")
	assert.Contains(t, content, "unique: a,b,c")
	assert.Contains(t, content, "combine: value1,value3")
}

// TestWorkerPool_Basic 测试 Worker 池基本功能
func TestWorkerPool_Basic(t *testing.T) {
	pool, err := template.NewWorkerPool(3, "../../scripts/render_worker.py")
	assert.NoError(t, err)
	defer pool.Close()

	assert.Equal(t, 3, pool.Size())
	assert.Equal(t, 3, pool.HealthCheck())

	// 创建临时模板
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "pool_test.yaml.j2")
	templateContent := `value: {{ num }}`
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	assert.NoError(t, err)

	// 多次渲染（使用不同的 workers）
	for i := 0; i < 5; i++ {
		content, err := pool.Render(template.RenderRequest{
			TemplatePath: templatePath,
			Context: map[string]interface{}{
				"num": i,
			},
		})

		assert.NoError(t, err)
		assert.Contains(t, content, "value:")
	}
}

// TestWorkerPool_Concurrent 测试并发渲染
func TestWorkerPool_Concurrent(t *testing.T) {
	pool, err := template.NewWorkerPool(5, "../../scripts/render_worker.py")
	assert.NoError(t, err)
	defer pool.Close()

	// 创建临时模板
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "concurrent_test.yaml.j2")
	templateContent := `result: {{ input }}`
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	assert.NoError(t, err)

	// 并发渲染
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(num int) {
			content, err := pool.Render(template.RenderRequest{
				TemplatePath: templatePath,
				Context: map[string]interface{}{
					"input": num,
				},
			})

			assert.NoError(t, err)
			assert.Contains(t, content, "result:")
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 检查健康状态
	assert.Equal(t, 5, pool.HealthCheck())
}

// TestHealthChecker_Basic 测试健康检查器
func TestHealthChecker_Basic(t *testing.T) {
	pool, err := template.NewWorkerPool(3, "../../scripts/render_worker.py")
	assert.NoError(t, err)
	defer pool.Close()

	checker := template.NewHealthChecker(pool, 5*time.Second)

	// 获取初始状态
	status := checker.GetStatus()
	assert.Equal(t, 3, status.TotalWorkers)
	assert.Equal(t, 3, status.AliveWorkers)
	assert.Equal(t, 0, status.DeadWorkers)
	assert.Equal(t, 1.0, status.HealthyRatio)
	assert.True(t, status.IsHealthy)

	// 启动健康检查
	checker.Start()
	defer checker.Stop()

	// 等待一小段时间
	time.Sleep(100 * time.Millisecond)

	// 再次检查状态
	status = checker.GetStatus()
	assert.True(t, status.IsHealthy)
}

// TestWorker_ErrorHandling 测试错误处理
func TestWorker_ErrorHandling(t *testing.T) {
	worker, err := template.NewPythonWorker("../../scripts/render_worker.py")
	assert.NoError(t, err)
	defer worker.Close()

	// 测试不存在的模板
	_, err = worker.Render(template.RenderRequest{
		TemplatePath: "/nonexistent/path/template.yaml.j2",
		Context:      map[string]interface{}{},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "模板文件不存在")
}
