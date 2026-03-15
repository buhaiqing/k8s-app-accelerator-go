package template

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type WorkerPool struct {
	workers    []*PythonWorker
	current    uint64
	mutex      sync.RWMutex
	scriptPath string
	size       int
	blacklist  map[int]time.Time
	blackMu    sync.RWMutex
}

func NewWorkerPool(count int, scriptPath string) (*WorkerPool, error) {
	if count <= 0 {
		count = 5
	}

	pool := &WorkerPool{
		workers:    make([]*PythonWorker, count),
		scriptPath: scriptPath,
		size:       count,
		blacklist:  make(map[int]time.Time),
	}

	for i := 0; i < count; i++ {
		worker, err := NewPythonWorker(scriptPath)
		if err != nil {
			pool.Close()
			return nil, fmt.Errorf("创建 Worker %d 失败：%w", i, err)
		}
		pool.workers[i] = worker
	}

	return pool, nil
}

func (p *WorkerPool) GetWorker() *PythonWorker {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	attempts := 0
	maxAttempts := p.size * 2

	for attempts < maxAttempts {
		idx := atomic.AddUint64(&p.current, 1) % uint64(p.size)
		worker := p.workers[idx]

		if worker != nil && worker.IsAlive() && !p.isBlacklisted(worker.PID()) {
			return worker
		}

		attempts++
	}

	for _, worker := range p.workers {
		if worker != nil && worker.IsAlive() {
			return worker
		}
	}

	return nil
}

func (p *WorkerPool) isBlacklisted(pid int) bool {
	p.blackMu.RLock()
	defer p.blackMu.RUnlock()

	if blacklistedAt, exists := p.blacklist[pid]; exists {
		if time.Since(blacklistedAt) > 30*time.Second {
			p.blackMu.RUnlock()
			p.blackMu.Lock()
			delete(p.blacklist, pid)
			p.blackMu.Unlock()
			p.blackMu.RLock()
			return false
		}
		return true
	}
	return false
}

func (p *WorkerPool) addToBlacklist(pid int) {
	p.blackMu.Lock()
	defer p.blackMu.Unlock()
	p.blacklist[pid] = time.Now()
}

func (p *WorkerPool) Render(req RenderRequest) (string, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		worker := p.GetWorker()
		if worker == nil {
			return "", fmt.Errorf("没有可用的 Worker")
		}

		content, err := worker.Render(req)
		if err != nil {
			lastErr = err
			p.addToBlacklist(worker.PID())

			if attempt < maxRetries-1 {
				time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
			}
			continue
		}

		return content, nil
	}

	return "", fmt.Errorf("渲染失败（已重试 %d 次）：%w", maxRetries, lastErr)
}

func (p *WorkerPool) Close() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, worker := range p.workers {
		if worker != nil {
			worker.Close()
		}
	}
}

func (p *WorkerPool) Size() int {
	return p.size
}

func (p *WorkerPool) HealthCheck() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	aliveCount := 0
	for _, worker := range p.workers {
		if worker != nil && worker.IsAlive() {
			aliveCount++
		}
	}
	return aliveCount
}

func (p *WorkerPool) RestartDeadWorkers() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var errors []error

	for i, worker := range p.workers {
		if worker == nil || !worker.IsAlive() {
			if worker != nil {
				worker.Close()
			}

			newWorker, err := NewPythonWorker(p.scriptPath)
			if err != nil {
				errors = append(errors, fmt.Errorf("重启 Worker %d 失败：%w", i, err))
			} else {
				p.workers[i] = newWorker
				p.blackMu.Lock()
				delete(p.blacklist, newWorker.PID())
				p.blackMu.Unlock()
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("重启 Workers 时遇到 %d 个错误：%v", len(errors), errors)
	}

	return nil
}

func (p *WorkerPool) GetStats() map[string]interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	aliveCount := 0
	for _, worker := range p.workers {
		if worker != nil && worker.IsAlive() {
			aliveCount++
		}
	}

	p.blackMu.RLock()
	blacklistCount := len(p.blacklist)
	p.blackMu.RUnlock()

	return map[string]interface{}{
		"total_workers":     p.size,
		"alive_workers":     aliveCount,
		"dead_workers":      p.size - aliveCount,
		"blacklisted_count": blacklistCount,
		"healthy_ratio":     float64(aliveCount) / float64(p.size),
	}
}
