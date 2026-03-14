package template

import (
	"fmt"
	"sync"
)

// WorkerPool Python Worker 进程池
type WorkerPool struct {
	workers    []*PythonWorker
	current    int
	mutex      sync.Mutex
	scriptPath string
	size       int
}

// NewWorkerPool 创建 Worker 池
// count: Worker 数量（推荐 5 个）
// scriptPath: render_worker.py 的路径
func NewWorkerPool(count int, scriptPath string) (*WorkerPool, error) {
	if count <= 0 {
		count = 5 // 默认 5 个 workers
	}

	pool := &WorkerPool{
		workers:    make([]*PythonWorker, count),
		current:    0,
		scriptPath: scriptPath,
		size:       count,
	}

	// 初始化所有 workers
	for i := 0; i < count; i++ {
		worker, err := NewPythonWorker(scriptPath)
		if err != nil {
			// 清理已创建的 workers
			pool.Close()
			return nil, fmt.Errorf("创建 Worker %d 失败：%w", i, err)
		}
		pool.workers[i] = worker
	}

	return pool, nil
}

// GetWorker 获取一个可用的 Worker
// 使用轮询方式分配 Worker，实现负载均衡
func (p *WorkerPool) GetWorker() *PythonWorker {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 轮询获取 worker
	worker := p.workers[p.current]
	p.current = (p.current + 1) % p.size

	return worker
}

// Render 使用 Worker 池渲染模板
// 自动获取可用 worker 并处理重试
func (p *WorkerPool) Render(req RenderRequest) (string, error) {
	worker := p.GetWorker()
	content, err := worker.Render(req)

	if err != nil {
		// 如果 worker 可能已失效，尝试其他 worker 重试
		for i := 0; i < p.size-1; i++ {
			worker = p.GetWorker()
			if worker.IsAlive() {
				content, err = worker.Render(req)
				if err == nil {
					return content, nil
				}
			}
		}
		return "", fmt.Errorf("所有 Worker 都失败：%w", err)
	}

	return content, nil
}

// Close 关闭所有 Workers
func (p *WorkerPool) Close() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, worker := range p.workers {
		if worker != nil {
			worker.Close()
		}
	}
}

// Size 获取 Worker 池大小
func (p *WorkerPool) Size() int {
	return p.size
}

// HealthCheck 健康检查
// 返回存活的 Worker 数量
func (p *WorkerPool) HealthCheck() int {
	aliveCount := 0
	for _, worker := range p.workers {
		if worker.IsAlive() {
			aliveCount++
		}
	}
	return aliveCount
}

// RestartDeadWorkers 重启已死亡的 Workers
func (p *WorkerPool) RestartDeadWorkers() error {
	var errors []error

	for i, worker := range p.workers {
		if !worker.IsAlive() {
			// 关闭旧 worker
			if err := worker.Close(); err != nil {
				errors = append(errors, err)
			}

			// 创建新 worker
			newWorker, err := NewPythonWorker(p.scriptPath)
			if err != nil {
				errors = append(errors, fmt.Errorf("重启 Worker %d 失败：%w", i, err))
			} else {
				p.workers[i] = newWorker
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("重启 Workers 时遇到 %d 个错误：%v", len(errors), errors)
	}

	return nil
}
