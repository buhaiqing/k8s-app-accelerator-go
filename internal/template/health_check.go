package template

import (
	"fmt"
	"time"
)

// HealthChecker 健康检查器
type HealthChecker struct {
	pool     *WorkerPool
	interval time.Duration
	stopChan chan struct{}
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(pool *WorkerPool, checkInterval time.Duration) *HealthChecker {
	if checkInterval <= 0 {
		checkInterval = 30 * time.Second // 默认 30 秒检查一次
	}

	return &HealthChecker{
		pool:     pool,
		interval: checkInterval,
		stopChan: make(chan struct{}),
	}
}

// Start 启动健康检查（后台 goroutine）
func (h *HealthChecker) Start() {
	go func() {
		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				h.checkAndRecover()
			case <-h.stopChan:
				return
			}
		}
	}()
}

// Stop 停止健康检查
func (h *HealthChecker) Stop() {
	close(h.stopChan)
}

// checkAndRecover 检查并恢复
func (h *HealthChecker) checkAndRecover() {
	aliveCount := h.pool.HealthCheck()
	totalCount := h.pool.Size()

	if aliveCount < totalCount {
		// 有 worker 死亡，尝试重启
		deadCount := totalCount - aliveCount
		fmt.Printf("[健康检查] 发现 %d 个 Worker 已死亡，正在重启...\n", deadCount)

		if err := h.pool.RestartDeadWorkers(); err != nil {
			fmt.Printf("[健康检查] 重启 Worker 失败：%v\n", err)
		} else {
			fmt.Printf("[健康检查] 成功重启 %d 个 Worker\n", deadCount)
		}
	}
}

// GetStatus 获取健康状态
func (h *HealthChecker) GetStatus() HealthStatus {
	aliveCount := h.pool.HealthCheck()
	totalCount := h.pool.Size()

	return HealthStatus{
		TotalWorkers:     totalCount,
		AliveWorkers:     aliveCount,
		DeadWorkers:      totalCount - aliveCount,
		HealthyRatio:     float64(aliveCount) / float64(totalCount),
		IsHealthy:        aliveCount == totalCount,
		CheckIntervalSec: h.interval.Seconds(),
	}
}

// HealthStatus 健康状态
type HealthStatus struct {
	TotalWorkers     int     `json:"total_workers"`
	AliveWorkers     int     `json:"alive_workers"`
	DeadWorkers      int     `json:"dead_workers"`
	HealthyRatio     float64 `json:"healthy_ratio"`
	IsHealthy        bool    `json:"is_healthy"`
	CheckIntervalSec float64 `json:"check_interval_sec"`
}
