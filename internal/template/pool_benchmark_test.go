package template_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buhaiqing/k8s-app-accelerator-go/internal/template"
)

func BenchmarkWorkerPool_GetWorker(b *testing.B) {
	pool, err := template.NewWorkerPool(5, "../../scripts/render_worker.py")
	if err != nil {
		b.Fatalf("创建 Worker Pool 失败：%v", err)
	}
	defer pool.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		worker := pool.GetWorker()
		if worker == nil {
			b.Errorf("获取 worker 失败")
		}
	}
}

func BenchmarkWorkerPool_Render(b *testing.B) {
	pool, err := template.NewWorkerPool(5, "../../scripts/render_worker.py")
	if err != nil {
		b.Fatalf("创建 Worker Pool 失败：%v", err)
	}
	defer pool.Close()

	// 创建临时模板
	tmpDir := b.TempDir()
	templatePath := filepath.Join(tmpDir, "benchmark.yaml.j2")
	templateContent := `benchmark: {{ value }}`
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		b.Fatalf("创建模板文件失败：%v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pool.Render(template.RenderRequest{
			TemplatePath: templatePath,
			Context: map[string]interface{}{
				"value": i,
			},
		})
		if err != nil {
			b.Errorf("渲染失败：%v", err)
		}
	}
}

func BenchmarkWorkerPool_RenderParallel(b *testing.B) {
	pool, err := template.NewWorkerPool(10, "../../scripts/render_worker.py")
	if err != nil {
		b.Fatalf("创建 Worker Pool 失败：%v", err)
	}
	defer pool.Close()

	// 创建临时模板
	tmpDir := b.TempDir()
	templatePath := filepath.Join(tmpDir, "benchmark_parallel.yaml.j2")
	templateContent := `parallel: {{ value }}`
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		b.Fatalf("创建模板文件失败：%v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, err := pool.Render(template.RenderRequest{
				TemplatePath: templatePath,
				Context: map[string]interface{}{
					"value": i,
				},
			})
			if err != nil {
				b.Errorf("渲染失败：%v", err)
			}
			i++
		}
	})
}

func BenchmarkWorkerPool_HealthCheck(b *testing.B) {
	pool, err := template.NewWorkerPool(5, "../../scripts/render_worker.py")
	if err != nil {
		b.Fatalf("创建 Worker Pool 失败：%v", err)
	}
	defer pool.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alive := pool.HealthCheck()
		if alive != 5 {
			b.Errorf("健康检查失败，期望 5，实际 %d", alive)
		}
	}
}
