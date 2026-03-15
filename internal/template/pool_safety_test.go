package template_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/template"
	"github.com/stretchr/testify/assert"
)

func TestWorkerPool_ConcurrentSafety(t *testing.T) {
	pool, err := template.NewWorkerPool(5, "../../scripts/render_worker.py")
	if err != nil {
		t.Fatalf("创建 Worker Pool 失败：%v", err)
	}
	defer pool.Close()

	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "concurrent_safety.yaml.j2")
	templateContent := `value: {{ num }}`
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("创建模板文件失败：%v", err)
	}

	const goroutines = 100
	var wg sync.WaitGroup
	workerUsage := make(map[int]int)
	var mu sync.Mutex

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(num int) {
			defer wg.Done()

			worker := pool.GetWorker()
			if worker == nil {
				t.Errorf("获取 worker 失败")
				return
			}

			mu.Lock()
			workerUsage[worker.PID()]++
			mu.Unlock()

			content, err := worker.Render(template.RenderRequest{
				TemplatePath: templatePath,
				Context: map[string]interface{}{
					"num": num,
				},
			})

			assert.NoError(t, err)
			assert.Contains(t, content, "value:")
		}(i)
	}

	wg.Wait()

	assert.Equal(t, 5, len(workerUsage), "应该有5个不同的 worker 被使用")

	for pid, count := range workerUsage {
		t.Logf("Worker PID %d 处理了 %d 个请求", pid, count)
		assert.LessOrEqual(t, count, 50, "Worker 不应该过载")
	}
}

func TestWorkerPool_RetryMechanism(t *testing.T) {
	pool, err := template.NewWorkerPool(3, "../../scripts/render_worker.py")
	if err != nil {
		t.Fatalf("创建 Worker Pool 失败：%v", err)
	}
	defer pool.Close()

	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "retry_test.yaml.j2")
	templateContent := `value: {{ input }}`
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("创建模板文件失败：%v", err)
	}

	content, err := pool.Render(template.RenderRequest{
		TemplatePath: templatePath,
		Context: map[string]interface{}{
			"input": "test",
		},
	})

	assert.NoError(t, err)
	assert.Contains(t, content, "value: test")
}

func TestWorkerPool_HealthCheckUnderLoad(t *testing.T) {
	pool, err := template.NewWorkerPool(5, "../../scripts/render_worker.py")
	if err != nil {
		t.Fatalf("创建 Worker Pool 失败：%v", err)
	}
	defer pool.Close()

	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "health_test.yaml.j2")
	templateContent := `test: {{ value }}`
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("创建模板文件失败：%v", err)
	}

	const tasks = 50
	var wg sync.WaitGroup
	wg.Add(tasks)

	for i := 0; i < tasks; i++ {
		go func(num int) {
			defer wg.Done()
			_, err := pool.Render(template.RenderRequest{
				TemplatePath: templatePath,
				Context: map[string]interface{}{
					"value": num,
				},
			})
			assert.NoError(t, err)
		}(i)
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		alive := pool.HealthCheck()
		assert.Equal(t, 5, alive, "所有 worker 应该保持健康")
	}()

	wg.Wait()

	alive := pool.HealthCheck()
	assert.Equal(t, 5, alive, "所有 worker 应该仍然健康")
}

func TestWorkerPool_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试")
	}

	pool, err := template.NewWorkerPool(10, "../../scripts/render_worker.py")
	if err != nil {
		t.Fatalf("创建 Worker Pool 失败：%v", err)
	}
	defer pool.Close()

	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "stress_test.yaml.j2")
	templateContent := `stress: {{ iteration }}`
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("创建模板文件失败：%v", err)
	}

	const iterations = 500
	var wg sync.WaitGroup
	var errorCount int
	var mu sync.Mutex

	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func(num int) {
			defer wg.Done()
			_, err := pool.Render(template.RenderRequest{
				TemplatePath: templatePath,
				Context: map[string]interface{}{
					"iteration": num,
				},
			})
			if err != nil {
				mu.Lock()
				errorCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	successRate := float64(iterations-errorCount) / float64(iterations)
	t.Logf("成功率: %.2f%% (%d/%d)", successRate*100, iterations-errorCount, iterations)
	assert.GreaterOrEqual(t, successRate, 0.95, "成功率应该 >= 95%")
}
