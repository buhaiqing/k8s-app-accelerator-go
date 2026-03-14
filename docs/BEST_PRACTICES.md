# 开发最佳实践

**前置条件**: 
- ✅ 已完成 [DEVELOPMENT_GUIDE.md](./DEVELOPMENT_GUIDE.md)
- ✅ 有实际开发经验
- ✅ 遇到具体问题

**目标读者**: 开发者  
**最后更新**: 2025-03-14

---

## 📝 Python 环境要求

### 版本要求

```bash
# 最低版本
Python >= 3.7

# 推荐版本
Python >= 3.9
```

### 依赖安装

```bash
# 安装依赖
pip3 install -r scripts/requirements.txt

# 验证安装
python3 -c "import jinja2; print(jinja2.__version__)"
python3 -c "import yaml; print(yaml.__version__)"
```

### 虚拟环境（推荐）

```bash
# 创建虚拟环境
python3 -m venv venv

# 激活虚拟环境
source venv/bin/activate  # macOS/Linux
venv\Scripts\activate     # Windows

# 安装依赖
pip3 install -r scripts/requirements.txt
```

---

## 💻 Go 版本要求

### 版本要求

```bash
# 最低版本
Go >= 1.21

# 推荐版本
Go >= 1.22
```

### 模块管理

```bash
# 下载依赖
go mod download

# 整理依赖
go mod tidy

# 添加新依赖
go get github.com/spf13/cobra
```

---

## 🛣️ 路径处理规范

### 使用 filepath 包

```go
// ✅ 正确做法 - 跨平台兼容
import "path/filepath"

path := filepath.Join("roles", roleName, "templates")
absPath, _ := filepath.Abs(path)
```

```go
// ❌ 错误做法 - 硬编码斜杠
path := "roles/" + roleName + "/templates"

// Windows 上会失败！
```

### 工作目录处理

```go
// 获取工作目录
workDir, err := os.Getwd()

// 转换为绝对路径
absPath, err := filepath.Abs(relativePath)

// 清理路径
cleanPath := filepath.Clean(dirtyPath)
```

---

## ⚠️ 错误处理模式

### 推荐的错误处理方式

```go
// ✅ 添加上下文信息
result, err := worker.Render(req)
if err != nil {
    return fmt.Errorf("render template failed: %w", err)
}

// ✅ 检查必填参数
if ctx.Profile == "" {
    return fmt.Errorf("profile is required for rendering")
}
```

### 错误包装链

```go
func GenerateConfig() error {
    config, err := loadConfig()
    if err != nil {
        return fmt.Errorf("load config failed: %w", err)
    }
    
    result, err := renderTemplate(config)
    if err != nil {
        return fmt.Errorf("render template failed: %w", err)
    }
    
    return nil
}
```

---

## 🐛 调试技巧

### 1. 查看详细日志

```bash
# 添加 -v flag
go run cmd/main.go generate -v

# 输出示例
[INFO] Loading config from configs/vars.yaml
[INFO] Pre-check passed
[INFO] Starting worker pool (5 workers)
[DEBUG] Rendering app.yaml.j2 for gateway-service
[INFO] Generated output/output/dly/production/k8s_zt4d/gateway-service.yaml
```

### 2. 单步调试

```bash
# 只生成一个应用
go run cmd/main.go generate \
  --base-dir configs \
  --roles gateway-service \
  --verbose
```

### 3. 测试单个 Worker

```bash
# 手动启动 Worker
python3 scripts/render_worker.py --worker-mode

# 发送测试请求
echo '{"template_path": "test.j2", "context": {"name": "test"}}' > /tmp/test.json
cat /tmp/test.json | python3 scripts/render_worker.py --worker-mode
```

---

## 🚀 性能优化技巧

### 1. 控制 Worker 数量

```go
// 推荐：5 个 Workers
pool := NewWorkerPool(5, scriptPath)

// 太多会导致资源竞争
pool := NewWorkerPool(20, scriptPath) // ❌ 不推荐
```

### 2. 避免重复渲染

```go
// ✅ 使用缓存
cache := make(map[string]string)

content, exists := cache[templateKey]
if !exists {
    content, _ = worker.Render(templatePath, ctx)
    cache[templateKey] = content
}
```

### 3. 并发控制

```go
// 限制最大并发数
semaphore := make(chan struct{}, 10)

for _, app := range apps {
    semaphore <- struct{}{}
    go func(app App) {
        defer func() { <-semaphore }()
        generateApp(app)
    }(app)
}
```

---

## 🧪 测试策略

### 单元测试

```go
func TestArgoCDGenerator(t *testing.T) {
    config := &config.ProjectConfig{
        Project: "test",
    }
    
    gen := NewArgoCDGenerator(config, "/tmp/output")
    err := gen.GenerateAll()
    
    if err != nil {
        t.Fatalf("Generate failed: %v", err)
    }
}
```

### 集成测试

```bash
#!/bin/bash
# tests/integration_test.sh

set -e

echo "Running integration test..."

# 清理输出
rm -rf /tmp/test-output

# 生成配置
go run cmd/main.go generate \
  --base-dir tests/fixtures \
  -o /tmp/test-output

# 验证输出文件存在
test -f /tmp/test-output/test.yaml || exit 1

echo "Integration test passed!"
```

---

## 📊 代码组织建议

### 1. 按功能分层

```
internal/
├── cli/          # CLI 命令层
├── config/       # 配置加载层
├── model/        # 数据模型层
├── template/     # 模板渲染层
├── generator/    # 配置生成层
└── validator/    # 配置校验层
```

### 2. 接口定义清晰

```go
// 定义接口
type Generator interface {
    GenerateAll() error
}

// 实现接口
type ArgoCDGenerator struct {
    // ...
}

func (g *ArgoCDGenerator) GenerateAll() error {
    // ...
}
```

### 3. 依赖注入

```go
// ✅ 推荐：通过构造函数注入依赖
func NewArgoCDGenerator(
    config *ProjectConfig,
    outputDir string,
) *ArgoCDGenerator {
    return &ArgoCDGenerator{
        projectConfig: config,
        outputDir:     outputDir,
    }
}

// ❌ 不推荐：全局变量
var globalConfig *ProjectConfig
```

---

## 🔒 安全注意事项

### 1. 敏感信息处理

```go
// ✅ 从环境变量读取
dbPassword := os.Getenv("DB_PASSWORD")

// ❌ 硬编码在代码中
dbPassword := "secret123"  // 不安全！
```

### 2. 文件权限控制

```go
// 设置合理的文件权限
os.WriteFile(path, content, 0644)  // 普通文件
os.WriteFile(path, content, 0755)  // 可执行文件
```

### 3. 输入验证

```go
// 验证项目名称格式
if !regexp.MustCompile(`^[a-z0-9]+$`).MatchString(project) {
    return fmt.Errorf("invalid project name")
}
```

---

## 📚 相关文档

- [DEVELOPMENT_GUIDE.md](./DEVELOPMENT_GUIDE.md) - 开发入门指南
- [ARCHITECTURE_DEEP_DIVE.md](./ARCHITECTURE_DEEP_DIVE.md) - 架构设计详解
- [PYTHON_WORKER_IMPLEMENTATION.md](./PYTHON_WORKER_IMPLEMENTATION.md) - Worker 实现细节

---

**最后更新**: 2025-03-14  
**维护者**: K8s App Accelerator Team
