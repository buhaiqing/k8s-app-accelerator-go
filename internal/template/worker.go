package template

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// RenderRequest 渲染请求
type RenderRequest struct {
	TemplatePath string                 `json:"template_path"`
	Context      map[string]interface{} `json:"context"`
}

// RenderResponse 渲染响应
type RenderResponse struct {
	Success bool   `json:"success"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

// PythonWorker Python Worker 封装
type PythonWorker struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	reader *bufio.Reader
	mutex  sync.Mutex
	pid    int
}

// NewPythonWorker 创建新的 Python Worker
func NewPythonWorker(scriptPath string) (*PythonWorker, error) {
	// 转换为绝对路径
	absPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("获取脚本绝对路径失败：%w", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(absPath); err != nil {
		return nil, fmt.Errorf("脚本文件不存在：%s: %w", absPath, err)
	}

	// 创建命令（以 worker 模式运行）
	cmd := exec.Command("python3", absPath, "--worker-mode")

	// 获取 stdin/stdout 管道
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("获取 stdin 失败：%w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("获取 stdout 失败：%w", err)
	}

	// 重定向 stderr 到主程序的 stderr（方便调试）
	cmd.Stderr = nil

	// 启动进程
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("启动 Worker 进程失败：%w", err)
	}

	worker := &PythonWorker{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		reader: bufio.NewReader(stdout),
		pid:    cmd.Process.Pid,
	}

	return worker, nil
}

// Render 渲染模板
func (w *PythonWorker) Render(req RenderRequest) (string, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// 编码请求为 JSON
	requestData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("编码请求失败：%w", err)
	}

	// 发送请求（添加换行符）
	requestData = append(requestData, '\n')
	_, err = w.stdin.Write(requestData)
	if err != nil {
		return "", fmt.Errorf("发送请求失败：%w", err)
	}

	// 读取响应（读取一行）
	responseLine, err := w.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("读取响应失败：%w", err)
	}

	// 解析响应
	var resp RenderResponse
	if err := json.Unmarshal([]byte(responseLine), &resp); err != nil {
		return "", fmt.Errorf("解析响应失败：%w", err)
	}

	// 检查是否成功
	if !resp.Success {
		return "", fmt.Errorf("渲染失败：%s", resp.Error)
	}

	return resp.Content, nil
}

// Close 关闭 Worker
func (w *PythonWorker) Close() error {
	// 关闭 stdin（会触发 Worker 退出）
	if w.stdin != nil {
		w.stdin.Close()
	}

	// 等待进程结束
	if w.cmd != nil && w.cmd.Process != nil {
		// 给进程一个优雅退出的机会
		done := make(chan error, 1)
		go func() {
			done <- w.cmd.Wait()
		}()

		// 等待最多 1 秒
		select {
		case <-done:
			// 正常退出
		case <-time.After(1 * time.Second):
			// 超时，强制退出
		}
	}

	return nil
}

// IsAlive 检查 Worker 是否存活
func (w *PythonWorker) IsAlive() bool {
	if w.cmd == nil || w.cmd.Process == nil {
		return false
	}

	// 检查进程是否已经结束
	// ProcessState != nil 表示进程已经结束
	return w.cmd.ProcessState == nil
}

// PID 获取进程 ID
func (w *PythonWorker) PID() int {
	return w.pid
}
