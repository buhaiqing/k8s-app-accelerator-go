# 开发指南

**前置条件**: 无  
**目标读者**: 新用户、开发者  
**最后更新**: 2025-03-14

---

## 🚀 快速开始

### 1. 环境准备

#### Go 环境

```bash
# 版本要求
Go >= 1.21

# 验证安装
go version

# 输出示例
go version go1.21.0 darwin/arm64
```

#### Python 环境

```bash
# 版本要求
Python >= 3.10

# 验证安装
python3 --version

# 输出示例
Python 3.10
```

#### 系统依赖

```bash
# macOS
brew install go python3

# Ubuntu/Debian
sudo apt-get install golang-go python3 python3-pip

# Windows (WSL)
sudo apt-get install golang-go python3 python3-pip
```

---

### 2. 安装依赖

#### Go 依赖

```bash
# 克隆项目
git clone https://github.com/buhaiqing/k8s-app-accelerator-go.git
cd k8s-app-accelerator-go

# 安装 Go 模块
go mod download
```

#### Python 依赖

```bash
# 进入项目目录
cd k8s-app-accelerator-go

# 安装 Python 包
pip3 install -r scripts/requirements.txt

# 验证安装
python3 -c "import jinja2; print(f'Jinja2 {jinja2.__version__}')"
```

---

### 3. 第一次运行

#### 预检配置

```bash
# 运行预检
go run cmd/main.go precheck --base-dir configs

# 查看详细日志
go run cmd/main.go precheck \
  --base-dir configs \
  --verbose
```

#### 生成配置

```bash
# 生成 K8s 配置
go run cmd/main.go generate --base-dir configs

# 生成 ArgoCD 配置
go run cmd/main.go argocd generate --base-dir configs

# 生成 Jenkins 配置
go run cmd/main.go jenkins generate --base-dir configs
```

---

## 🛠️ Makefile 使用

### 常用命令

```makefile
# 生成所有 K8s 配置
make generate

# 生成 ArgoCD 配置
make argocd

# 生成 Jenkins 配置
make jenkins

# 生成 CMDB SQL
make cmdb

# 预检配置
make precheck

# 运行测试
make test

# 清理输出
make clean

# 编译发布版本
make build
```

### 自定义构建

```makefile
# 编译 Linux 版本
CGO_ENABLED=0 GOOS=linux go build -o k8s-gen-linux cmd/main.go

# 编译 macOS 版本
CGO_ENABLED=0 GOOS=darwin go build -o k8s-gen-darwin cmd/main.go

# 编译 Windows 版本
CGO_ENABLED=0 GOOS=windows go build -o k8s-gen-windows.exe cmd/main.go
```

---

## 📁 项目结构

```
k8s-app-accelerator-go/
├── cmd/                    # CLI 入口
│   └── main.go
├── internal/               # 内部包
│   ├── cli/               # CLI 命令实现
│   ├── config/            # 配置加载
│   ├── model/             # 数据模型
│   ├── template/          # 模板渲染
│   ├── generator/         # 配置生成
│   └── validator/         # 配置校验
├── scripts/                # Python 脚本
│   ├── render_worker.py   # Jinja2 渲染 Worker
│   ├── filters.py         # Ansible filters
│   └── requirements.txt   # Python 依赖
├── templates/              # Jinja2 模板
│   ├── argo-app/          # ArgoCD Application
│   └── jenkins-jobs/      # Jenkins Jobs
├── configs/                # 配置文件
│   ├── vars.yaml
│   ├── resources.yaml
│   └── mapping.yaml
├── output/                 # 生成结果
├── Makefile
├── go.mod
└── README.md
```

---

## 🔧 调试技巧

### 1. 查看详细日志

```bash
# 添加 -v flag
go run cmd/main.go generate \
  --base-dir configs \
  --verbose
```

### 2. 单步执行

```bash
# 只生成一个应用
go run cmd/main.go generate \
  --base-dir configs \
  --roles gateway-service
```

### 3. 检查 Python Worker

```bash
# 手动测试 Worker
python3 scripts/render_worker.py --worker-mode

# 测试 Filters
python3 -c "from scripts.filters import *; print(ternary(True, 'yes', 'no'))"
```

### 4. 验证输出

```bash
# 对比 Ansible 和 Go 输出
diff -r output/go output/ansible

# 使用对比脚本（统一对比工具）
bash scripts/compare_harness.sh
```

---

## ⚡ 性能优化

### 1. 并发控制

默认使用 5 个 Python Worker 并发渲染：

```go
// internal/template/python_pool.go
pool := NewWorkerPool(5, "scripts/render_worker.py")
```

### 2. 缓存机制

渲染结果会自动缓存，避免重复计算。

### 3. 超时设置

单次渲染超时 30 秒自动失败：

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

---

## 🐛 常见问题

### Q1: Python Worker 启动失败

**症状**: `exec: "python3": executable file not found`

**解决**:
```bash
# 检查 Python 路径
which python3

# 创建软链接
sudo ln -s $(which python3) /usr/local/bin/python3
```

### Q2: Jinja2 模板渲染失败

**症状**: `template not found`

**解决**:
```bash
# 检查模板路径
ls -la templates/

# 使用绝对路径
export TEMPLATE_DIR=$(pwd)/templates
```

### Q3: YAML 解析错误

**症状**: `yaml: line X: did not find expected key`

**解决**:
```bash
# 验证 YAML 格式
python3 -c "import yaml; yaml.safe_load(open('configs/vars.yaml'))"

# 检查缩进
cat -A configs/vars.yaml
```

### Q4: 权限问题

**症状**: `permission denied`

**解决**:
```bash
# 赋予执行权限
chmod +x scripts/render_worker.py

# 检查输出目录权限
chmod -R 755 output/
```

---

## 📚 下一步

完成本指南后，建议继续阅读：

1. **[CLI_REFERENCE.md](./CLI_REFERENCE.md)** - 详细的命令参考
2. **[ARCHITECTURE_DEEP_DIVE.md](./ARCHITECTURE_DEEP_DIVE.md)** - 深入理解架构设计
3. **[BEST_PRACTICES.md](./BEST_PRACTICES.md)** - 开发最佳实践

---

## 🔗 相关资源

- [Go 官方文档](https://golang.org/doc/)
- [Python 官方文档](https://docs.python.org/3/)
- [Jinja2 官方文档](https://jinja.palletsprojects.com/)
- [Makefile 教程](https://www.gnu.org/software/make/manual/)

---

**最后更新**: 2025-03-14  
**维护者**: K8s App Accelerator Team
