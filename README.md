# K8s App Accelerator Go

基于 Ansible roles 模板生成 Kubernetes 应用配置的 Golang实现。

## 🎯 项目目标

- ✅ **100% 兼容现有 Jinja2 模板**（无需修改 Ansible roles）
- ✅ **性能提升 5 倍以上**（从 8 分钟缩短到 1.5 分钟）
- ✅ **跨平台支持**（Windows/macOS/Linux原生运行）
- ✅ **智能预检功能**（减少 80% 配置错误）
- ✅ **开发友好**（支持 `go run` 直接运行）

## 📦 快速开始

### 安装依赖

```bash
go mod download
pip3 install -r scripts/requirements.txt
```

### 构建

```bash
# 开发模式（推荐）
go run main.go --help

# 编译二进制
go build -o k8s-gen
```

### 使用示例

```bash
# 预检配置
go run main.go precheck --base-dir /path/to/configs

# 生成配置（最简单的方式）
go run main.go generate --base-dir /path/to/configs

# 使用默认 base-dir
go run main.go generate
```

## 🏗️ 项目结构

```
k8s-app-accelerator-go/
├── internal/
│   ├── cli/           # CLI 命令入口
│   ├── config/        # 配置加载层
│   └── ...            # 其他模块（逐步实现）
├── configs/           # 示例配置文件
├── scripts/           # Python 脚本（Jinja2 Worker）
└── main.go           # 程序入口
```

## 📝 配置文件说明

### 1. vars.yaml - 项目配置

```yaml
project: my-project
profiles:
  - int
  - uat
  - production
apollo:
  site: https://apollo.example.com
  token: your-token
```

### 2. resources.yaml - 资源组配置

```yaml
rds:
  - name: default
    datasource_url: rm-xxxxx.mysql.rds.aliyuncs.com
redis:
  - name: default
    redisIp: r-xxxxx.redis.rds.aliyuncs.com
```

### 3. mapping.yaml - 映射配置

```yaml
mappings:
  cms-service: cms
  fms-service: fms
```

## 🧪 测试

```bash
# 运行所有测试
go test -v ./...

# 运行特定包测试
go test -v ./internal/config/...
```

## 🚀 开发工作流

### Day 1-2: 项目骨架 + 配置加载层 ✅

- [x] 初始化 Go 模块
- [x] 添加 Cobra CLI 框架
- [x] 创建配置数据结构
- [x] 实现配置加载器
- [x] 编写单元测试

### Day 3-4: Python Worker + 进程池

- [x] 实现 render_worker.py
- [x] 实现 PythonWorker 封装
- [x] 实现 WorkerPool 进程池
- [x] 添加健康检查机制

### Day 5-6: Pre-Check 预检功能

- [ ] 实现配置校验器
- [ ] 实现预检规则
- [ ] 添加彩色输出报告

### Day 7-8: Generate 生成功能

- [ ] 实现上下文构建器
- [ ] 实现 Role 生成器
- [ ] 实现文件写入器

### Day 9-10: 测试 + 文档

- [ ] 集成测试
- [ ] 性能基准测试
- [ ] 完善文档

## 🛠️ 技术栈

- **语言**: Go 1.21+
- **CLI 框架**: Cobra
- **YAML 解析**: gopkg.in/yaml.v3
- **测试框架**: testify
- **模板引擎**: Jinja2 (Python Worker)

## 📋 下一步

1. 复制示例配置文件为实际使用的文件名
2. 填入实际的配置值
3. 等待后续功能实现完成
4. 运行 `precheck` 和 `generate` 命令

## 📚 更多文档

详细技术方案请参考 [AGENTS.md](./AGENTS.md)
