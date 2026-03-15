# 开发最佳实践

**前置条件**: 
- ✅ 已完成 [DEVELOPMENT_GUIDE.md](./DEVELOPMENT_GUIDE.md)
- ✅ 有实际开发经验
- ✅ 遇到具体问题

**目标读者**: 开发者  
**最后更新**: 2026-03-15

---

## 🎯 核心原则

### 1. 测试驱动开发（TDD）
- ✅ **先写测试，再写代码**
- ✅ **小步快跑，持续重构**
- ✅ **测试覆盖率 >= 70%**

### 2. 防御性编程
- ✅ **永远不要相信输入**
- ✅ **所有错误都要处理**
- ✅ **所有资源都要释放**

### 3. 简洁明了
- ✅ **代码是写给人看的**
- ✅ **复杂逻辑要注释**
- ✅ **命名要有意义**

---

## 🧪 持续测试与质量保障

### 测试作为质量保证的核心环节

单元测试是软件质量保证的关键环节，必须给予充分重视：

**1. 测试即文档**
- 好的测试用例是代码的活文档
- 明确表达业务需求和预期行为
- 帮助新成员快速理解代码逻辑

**2. 测试即防护网**
- 防止代码退化（Regression）
- 在重构时提供安全保障
- 快速发现引入的错误

**3. 测试即设计工具**
- TDD 驱动更好的代码设计
- 促进模块化和解耦
- 提高代码可维护性

**质量门槛**：
- ✅ **所有新功能必须包含测试**
- ✅ **测试覆盖率 >= 70%**
- ✅ **必须通过所有单元测试才能提交**
- ✅ **禁止提交会导致测试失败的代码**

---

### 使用 Air 实现持续测试

[Air](https://github.com/air-verse/air) 是一个 Go 语言的 live-reloading 工具，可以在文件变更时自动触发测试。

#### 安装 Air

```bash
# 方法1：通过 go install（推荐，Go 1.25+）
go install github.com/air-verse/air@latest

# 方法2：通过 Homebrew（macOS）
brew install go-air

# 方法3：通过 curl
curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

#### 配置 Air

在项目根目录创建 `.air.toml` 配置文件：

```toml
# .air.toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
# 监听以下文件变化
include_ext = ["go", "tpl", "tmpl", "html"]
# 忽略以下目录
exclude_dir = ["assets", "tmp", "vendor", "testdata"]
# 构建命令（这里运行测试）
cmd = "go test -v -race ./..."
# 立即执行一次
full_bin = "go test -v -race ./..."

[log]
# 日志级别
level = "info"

[misc]
# 退出时清理
clean_on_exit = true
```

#### 使用 Air 进行持续测试

```bash
# 启动持续测试（监听文件变更，自动运行测试）
air

# 或者指定配置文件
air -c .air.toml

# 查看帮助
air -h
```

**工作流程**：
1. 启动 `air` 后，Air 会监听项目文件变化
2. 当检测到 `.go` 文件被修改时，自动执行 `go test`
3. 测试结果实时显示在终端
4. 开发人员可以立即看到测试是否通过

#### Air 高级配置

```toml
[build]
# 监听多个扩展名
include_ext = "go, tpl, tmpl, html, yaml, yml, json, md"

# 排除特定目录
exclude_dir = "assets, tmp, vendor, testdata, build, dist"

# 构建超时
build.timeout = "10s"

# 重试次数
build.retry = 3

# 自定义测试命令
cmd = "go test -v -race -coverprofile=coverage.out ./..."

[screen]
# 保持滚动
clear_on_rebuild = false
keep_scroll = true

[misc]
# 按 Enter 重新运行
run_on_first_load = true
```

#### 完整的开发工作流

```bash
# 1. 初始化项目（如需要）
air init

# 2. 启动持续测试
air

# 3. 开发过程中：
#    - 修改代码 → Air 自动运行测试
#    - 查看测试结果
#    - 修复失败的测试

# 4. 测试全部通过后提交代码
git add .
git commit -m "feat: 添加新功能，测试通过"
```

---

### 测试失败时的处理流程

当监测到测试失败时，必须遵循以下处理流程：

#### 1. 立即停止开发

```bash
# 测试失败时的终端输出示例
=== RUN   TestWorkerPool_ConcurrentSafety
--- FAIL: TestWorkerPool_ConcurrentSafety (0.05s)
    pool_safety_test.go:87: Expected worker count to be 5, got 3

# ❌ 禁止继续开发
# ❌ 禁止提交失败的代码
# ✅ 必须先修复测试
```

#### 2. 分析失败原因

**常见失败原因**：

| 类型 | 描述 | 处理方式 |
|------|------|----------|
| 业务逻辑错误 | 代码实现不符合需求 | 修复业务逻辑 |
| 测试用例错误 | 测试本身有问题 | 更新测试用例 |
| 依赖问题 | 依赖版本不兼容 | 更新依赖 |
| 环境问题 | 环境配置错误 | 修复环境 |

**分析步骤**：
```bash
# 1. 查看详细错误信息
go test -v ./...

# 2. 运行竞态检测
go test -race ./...

# 3. 查看测试覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 4. 单独运行失败的测试
go test -v -run TestWorkerPool_ConcurrentSafety ./internal/template
```

#### 3. 修复代码或测试

**原则**：
- ✅ 如果是代码错误 → 修复代码
- ✅ 如果是测试用例错误 → 更新测试用例
- ✅ 如果需求变更 → 同步更新代码和测试
- ❌ 禁止跳过测试或删除测试

**修复示例**：

```go
// ❌ 错误做法：跳过测试
func TestWorkerPool_ConcurrentSafety(t *testing.T) {
    t.Skip("暂时跳过")  // ❌ 禁止这样做！
}

// ✅ 正确做法：修复问题
func TestWorkerPool_ConcurrentSafety(t *testing.T) {
    pool, err := template.NewWorkerPool(5, "script")
    require.NoError(t, err)
    
    // 测试逻辑...
    
    // 如果测试失败，分析原因并修复
    // 如果是代码问题，修复代码
    // 如果是测试问题，更新测试
}
```

#### 4. 验证修复

```bash
# 1. 运行所有测试
go test -v ./...

# 2. 运行竞态检测
go test -race ./...

# 3. 运行覆盖率
go test -cover ./...

# 4. 确保所有测试通过
```

#### 5. 提交代码

```bash
# 只有所有测试通过后才能提交
git add .
git commit -m "fix: 修复并发安全问题，测试通过"

# 验证 CI 状态
git status
```

---

### 持续集成要求

**本地提交前检查**：

```bash
# ✅ 必须全部通过
go test -race ./...           # 竞态检测
go test -cover ./...         # 覆盖率检查
go vet ./...                  # 静态分析
golangci-lint run            # 代码规范检查
```

**CI/CD 检查项**：

| 检查项 | 命令 | 要求 |
|--------|------|------|
| 单元测试 | `go test -race ./...` | 100% 通过 |
| 测试覆盖率 | `go test -cover ./...` | >= 70% |
| 代码规范 | `golangci-lint run` | 无警告 |
| 安全检查 | `go vet ./...` | 无漏洞 |

---

### 测试规范检查清单

```markdown
## 提交前必须检查

### 本地检查
- [ ] `go test -race ./...` 无竞态条件
- [ ] `go test -cover ./...` 覆盖率 >= 70%
- [ ] `go vet ./...` 无警告
- [ ] `golangci-lint run` 无错误

### 代码质量
- [ ] 所有新功能都有对应测试
- [ ] 测试用例描述清晰
- [ ] 测试用例可重复执行
- [ ] 测试用例相互独立

### 提交规范
- [ ] 所有测试通过后才能提交
- [ ] 提交信息描述清晰
- [ ] 无残留调试代码
- [ ] 代码符合规范
```

---

## ⚠️ 常见错误与反思

### 错误1：双入口文件混乱

#### 🔍 问题表现
```
/main.go          → 调用 cli.Execute()
/cmd/main.go      → 调用 cli.Execute()（无错误处理）
```

#### 💭 根本原因分析

**1. 缺乏项目结构规范**
- 没有明确的目录结构约定
- 不清楚 `cmd/` 目录的正确用途
- 缺少代码审查流程

**2. 对 Go 项目结构理解不足**
- `cmd/` 目录应该用于**子命令**，而非单一 CLI
- 单一 CLI 工具应该使用根目录的 `main.go`
- 缺少对 Go 标准项目布局的学习

**3. 缺少代码审查机制**
- 重复代码没有被及时发现
- 没有建立 PR 审查规范
- 缺少自动化检查工具

#### ✅ 正确做法

**项目结构规范**:
```
k8s-app-accelerator-go/
├── main.go                 # ✅ 单一 CLI 入口
├── internal/               # ✅ 内部包
│   ├── cli/               # CLI 命令
│   ├── config/            # 配置加载
│   ├── generator/         # 生成器
│   ├── template/          # 模板渲染
│   └── validator/         # 校验器
├── scripts/               # 脚本文件
├── docs/                  # 文档
└── go.mod                 # Go 模块定义
```

**何时使用 `cmd/` 目录**:
```go
// ✅ 正确：多个子命令
cmd/
├── generator/      # go run ./cmd/generator
│   └── main.go
├── validator/      # go run ./cmd/validator
│   └── main.go
└── cli/           # go run ./cmd/cli
    └── main.go

// ✅ 正确：单一 CLI 工具
main.go            # go run main.go
```

**代码审查清单**:
- [ ] 是否有重复的入口文件？
- [ ] 目录结构是否符合 Go 标准？
- [ ] 是否有冗余代码？
- [ ] 是否缺少错误处理？

#### 📋 规范化建议

**1. 建立项目结构文档**
```markdown
# 项目结构规范

## 单一 CLI 工具
- 使用根目录的 `main.go`
- 不创建 `cmd/` 目录

## 多命令工具
- 每个子命令放在 `cmd/<command>/main.go`
- 共享代码放在 `internal/`

## 禁止事项
- ❌ 不要创建重复的入口文件
- ❌ 不要在 `cmd/` 中放置共享代码
```

**2. 使用静态检查工具**
```bash
# 安装 golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行检查
golangci-lint run

# 配置 .golangci.yml
linters:
  enable:
    - dupl        # 检测重复代码
    - gofmt       # 格式检查
    - goimports   # 导入检查
    - govet       # 静态分析
```

**3. PR 审查规范**
```markdown
## PR 审查清单

### 代码质量
- [ ] 无重复代码
- [ ] 无冗余文件
- [ ] 符合 Go 标准

### 结构规范
- [ ] 目录结构正确
- [ ] 入口文件唯一
- [ ] 包划分合理

### 测试覆盖
- [ ] 单元测试通过
- [ ] 覆盖率 >= 70%
- [ ] 集成测试通过
```

---

### 错误2：Worker Pool 并发安全隐患

#### 🔍 问题表现

```go
// ❌ 问题代码
func (p *WorkerPool) GetWorker() *PythonWorker {
    p.mutex.Lock()
    defer p.mutex.Unlock()
    
    // 轮询分配 - 高并发下可能过载
    worker := p.workers[p.current]
    p.current = (p.current + 1) % p.size
    
    return worker
}

// ❌ 重试逻辑遍历所有 worker
for i := 0; i < p.size-1; i++ {
    worker = p.GetWorker()
    if worker.IsAlive() {
        content, err = worker.Render(req)
        if err == nil {
            return content, nil
        }
    }
}
```

#### 💭 根本原因分析

**1. 对并发编程理解不够深入**
- 不了解原子操作的优势
- 锁粒度太大（使用互斥锁而非读写锁）
- 没有考虑高并发场景下的性能

**2. 缺少并发安全测试**
- 没有并发压力测试
- 没有竞态条件检测
- 缺少负载均衡验证

**3. 缺少失败隔离机制**
- 失败的 worker 没有被隔离
- 没有黑名单机制
- 缺少自动恢复机制

**4. 缺少可观测性**
- 没有 Worker 统计信息
- 缺少健康监控
- 无法追踪负载分布

#### ✅ 正确做法

**1. 使用原子操作优化轮询**
```go
import "sync/atomic"

type WorkerPool struct {
    workers    []*PythonWorker
    current    uint64         // ✅ 使用 uint64 支持原子操作
    mutex      sync.RWMutex   // ✅ 使用读写锁
    size       int
}

func (p *WorkerPool) GetWorker() *PythonWorker {
    p.mutex.RLock()  // ✅ 读锁，允许多个并发读
    defer p.mutex.RUnlock()
    
    // ✅ 原子操作，无锁轮询
    idx := atomic.AddUint64(&p.current, 1) % uint64(p.size)
    worker := p.workers[idx]
    
    if worker != nil && worker.IsAlive() {
        return worker
    }
    
    // 降级：查找第一个可用的 worker
    for _, w := range p.workers {
        if w != nil && w.IsAlive() {
            return w
        }
    }
    
    return nil
}
```

**2. 实现失败隔离机制**
```go
type WorkerPool struct {
    blacklist  map[int]time.Time  // ✅ 黑名单
    blackMu    sync.RWMutex       // ✅ 黑名单锁
}

// 失败的 worker 加入黑名单
func (p *WorkerPool) addToBlacklist(pid int) {
    p.blackMu.Lock()
    defer p.blackMu.Unlock()
    p.blacklist[pid] = time.Now()
}

// 检查是否在黑名单中（30秒后自动恢复）
func (p *WorkerPool) isBlacklisted(pid int) bool {
    p.blackMu.RLock()
    defer p.blackMu.RUnlock()
    
    if blacklistedAt, exists := p.blacklist[pid]; exists {
        if time.Since(blacklistedAt) > 30*time.Second {
            // 自动恢复
            go func() {
                p.blackMu.Lock()
                delete(p.blacklist, pid)
                p.blackMu.Unlock()
            }()
            return false
        }
        return true
    }
    return false
}
```

**3. 实现指数退避重试**
```go
func (p *WorkerPool) Render(req RenderRequest) (string, error) {
    const maxRetries = 3
    var lastErr error
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        worker := p.GetWorker()
        if worker == nil {
            return "", fmt.Errorf("no available worker")
        }
        
        content, err := worker.Render(req)
        if err != nil {
            lastErr = err
            p.addToBlacklist(worker.PID())
            
            // ✅ 指数退避
            if attempt < maxRetries-1 {
                time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
            }
            continue
        }
        
        return content, nil
    }
    
    return "", fmt.Errorf("render failed after %d retries: %w", maxRetries, lastErr)
}
```

**4. 添加监控统计**
```go
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
```

#### 📋 规范化建议

**1. 并发编程检查清单**
```markdown
## 并发编程规范

### 锁的选择
- [ ] 读多写少 → 使用 `sync.RWMutex`
- [ ] 写多读少 → 使用 `sync.Mutex`
- [ ] 简单计数 → 使用 `atomic` 操作

### 原子操作优先
- [ ] 计数器 → `atomic.AddUint64`
- [ ] 标志位 → `atomic.Bool` (Go 1.19+)
- [ ] 指针 → `atomic.Value`

### 避免常见陷阱
- [ ] 不要在锁内调用外部函数
- [ ] 不要忘记释放锁（使用 defer）
- [ ] 不要复制锁（使用指针）
```

**2. 并发测试规范**
```go
// ✅ 必须编写并发测试
func TestWorkerPool_ConcurrentSafety(t *testing.T) {
    pool := NewWorkerPool(5)
    
    const goroutines = 100
    var wg sync.WaitGroup
    wg.Add(goroutines)
    
    for i := 0; i < goroutines; i++ {
        go func() {
            defer wg.Done()
            worker := pool.GetWorker()
            // 验证 worker 不为 nil
            if worker == nil {
                t.Error("worker is nil")
            }
        }()
    }
    
    wg.Wait()
}

// ✅ 使用竞态检测
// go test -race ./...
```

**3. 性能基准测试**
```go
func BenchmarkWorkerPool_GetWorker(b *testing.B) {
    pool := NewWorkerPool(5)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        pool.GetWorker()
    }
}

// 运行基准测试
// go test -bench=. -benchmem
```

---

### 错误3：Generator 错误处理不健壮

#### 🔍 问题表现

```go
// ❌ 问题代码
func (g *Generator) GenerateAll() error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(g.roleVars))
    
    for _, roleVars := range g.roleVars {
        wg.Add(1)
        go func(rv *model.RoleVars) {
            defer wg.Done()
            if err := g.generateForApp(rv); err != nil {
                errChan <- err
            }
        }(roleVars)
    }
    
    wg.Wait()
    close(errChan)
    
    // ❌ 只返回第一个错误
    if len(errChan) > 0 {
        return <-errChan
    }
    
    return nil
}
```

#### 💭 根本原因分析

**1. 对错误处理重视不够**
- 认为返回第一个错误就够了
- 没有考虑错误信息的完整性
- 缺少错误上下文

**2. 缺少并发控制意识**
- 不知道 goroutine 数量需要限制
- 没有考虑资源耗尽问题
- 缺少超时机制

**3. 没有使用标准库提供的并发工具**
- 不知道 `errgroup` 的存在
- 手动管理 goroutine 容易出错
- 缺少上下文传播

**4. 缺少上下文管理**
- 没有支持取消操作
- 无法优雅停止
- 缺少超时控制

#### ✅ 正确做法

**1. 使用 errgroup 管理并发**
```go
import (
    "context"
    "golang.org/x/sync/errgroup"
)

func (gen *Generator) GenerateAllWithContext(ctx context.Context) error {
    // ✅ 使用 errgroup 管理并发
    eg, ctx := errgroup.WithContext(ctx)
    
    // ✅ 限制并发数
    eg.SetLimit(10)
    
    // ✅ 收集所有错误
    var errors []error
    var errorMu sync.Mutex
    
    for _, roleVars := range gen.roleVars {
        rv := roleVars  // 捕获变量
        
        eg.Go(func() error {
            select {
            case <-ctx.Done():
                return ctx.Err()  // ✅ 支持取消
            default:
                if err := gen.generateForApp(rv); err != nil {
                    // ✅ 收集错误
                    errorMu.Lock()
                    errors = append(errors, err)
                    errorMu.Unlock()
                    return err
                }
                return nil
            }
        })
    }
    
    // ✅ 等待所有任务完成
    if err := eg.Wait(); err != nil {
        if len(errors) > 0 {
            return fmt.Errorf("generate failed (%d errors): %v", len(errors), errors[0])
        }
        return err
    }
    
    return nil
}
```

**2. 支持超时控制**
```go
func (gen *Generator) GenerateAllWithTimeout(timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    return gen.GenerateAllWithContext(ctx)
}

// 使用示例
err := gen.GenerateAllWithTimeout(5 * time.Minute)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Error("生成超时")
    }
    return err
}
```

**3. 错误信息结构化**
```go
type GenerateError struct {
    AppName string
    Stage   string
    Err     error
}

func (e *GenerateError) Error() string {
    return fmt.Sprintf("[%s] %s failed: %v", e.AppName, e.Stage, e.Err)
}

func (e *GenerateError) Unwrap() error {
    return e.Err
}

// 使用
if err := gen.generateForApp(rv); err != nil {
    return &GenerateError{
        AppName: rv.App,
        Stage:   "generate",
        Err:     err,
    }
}
```

#### 📋 规范化建议

**1. 并发控制规范**
```markdown
## 并发控制规范

### 必须限制并发数
- ✅ 使用信号量：`sem := make(chan struct{}, 10)`
- ✅ 使用 errgroup：`eg.SetLimit(10)`
- ❌ 不要无限制创建 goroutine

### 必须支持取消
- ✅ 所有长时间运行的任务都要支持 context
- ✅ 使用 `select { case <-ctx.Done(): }`
- ❌ 不要忽略取消信号

### 必须收集所有错误
- ✅ 使用 `[]error` 收集所有错误
- ✅ 使用 `sync.Mutex` 保护错误切片
- ❌ 不要只返回第一个错误
```

**2. 错误处理规范**
```go
// ✅ 错误要添加上下文
if err := doSomething(); err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// ✅ 使用自定义错误类型
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed: %s - %s", e.Field, e.Message)
}

// ✅ 错误要可追溯
if err := validate(config); err != nil {
    return fmt.Errorf("validate config %s failed: %w", config.Name, err)
}
```

**3. 资源管理规范**
```go
// ✅ 使用 defer 确保资源释放
file, err := os.Open(path)
if err != nil {
    return err
}
defer file.Close()

// ✅ 使用 context 控制超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// ✅ 使用 errgroup 管理 goroutine
eg, ctx := errgroup.WithContext(ctx)
defer eg.Wait()
```

---

## 📊 代码质量检查清单

### 提交前检查

```markdown
## 代码质量检查清单

### 并发安全
- [ ] 所有共享变量都有适当的锁保护
- [ ] 使用原子操作优化简单计数
- [ ] 所有 goroutine 都能正确退出
- [ ] 没有竞态条件（go test -race）

### 错误处理
- [ ] 所有错误都被处理
- [ ] 错误信息包含上下文
- [ ] 使用 fmt.Errorf 包装错误
- [ ] 资源都被正确释放

### 性能优化
- [ ] 使用读写锁优化读多写少场景
- [ ] 避免不必要的内存分配
- [ ] 使用 sync.Pool 复用对象
- [ ] 添加性能基准测试

### 测试覆盖
- [ ] 单元测试覆盖率 >= 70%
- [ ] 并发测试通过
- [ ] 压力测试通过
- [ ] 基准测试通过
```

---

## 🛠️ 工具推荐

### 静态检查工具

```bash
# golangci-lint - 综合检查工具
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 配置文件 .golangci.yml
linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - ineffassign
    - typecheck
    - gosimple
    - goconst
    - gocyclo
    - dupl
    - misspell

# 运行检查
golangci-lint run
```

### 竞态检测

```bash
# 运行测试时启用竞态检测
go test -race ./...

# 检测结果示例
==================
WARNING: DATA RACE
Write at 0x00c0000a6018 by goroutine 8:
  previous write at 0x00c0000a6018 by goroutine 7:
==================
```

### 性能分析

```bash
# CPU 性能分析
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# 内存分析
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# 查看火焰图
go tool pprof -http=:8080 cpu.prof
```

---

## 📚 学习资源

### 并发编程

1. **Go 并发编程实战** - 深入理解 goroutine 和 channel
2. **Go 内存模型** - 理解 happens-before 关系
3. **Advanced Patterns** - errgroup, sync.Pool, atomic

### 错误处理

1. **Go 错误处理最佳实践** - error wrapping 和自定义错误
2. **错误处理模式** - sentinel errors, custom errors
3. **错误追踪** - 使用 stack trace 和 context

### 性能优化

1. **Go 性能优化指南** - CPU、内存、I/O 优化
2. **pprof 使用指南** - 性能分析和调优
3. **基准测试** - 编写有效的 benchmark

---

## 🎓 总结

### 核心教训

**1. 项目结构要规范**
- 遵循 Go 标准项目布局
- 单一 CLI 使用根目录 main.go
- 多命令使用 cmd/ 目录

**2. 并发编程要谨慎**
- 使用原子操作优化简单计数
- 使用读写锁优化读多写少
- 实现失败隔离和自动恢复
- 添加监控和统计

**3. 错误处理要完整**
- 使用 errgroup 管理并发
- 收集所有错误信息
- 支持上下文取消
- 添加错误上下文

### 持续改进

**短期（1-2周）**
- ✅ 建立代码审查流程
- ✅ 配置静态检查工具
- ✅ 提高测试覆盖率

**中期（1个月）**
- 🟢 建立性能基准
- 🟢 实现监控指标
- 🟢 完善错误追踪

**长期（3个月）**
- 🔵 建立最佳实践库
- 🔵 定期代码审查
- 🔵 持续优化改进

---

**最后更新**: 2026-03-15  
**维护者**: K8s App Accelerator Team  
**版本**: v2.0（基于实战反思）
