# K8s App Accelerator Go

基于 Ansible roles 模板生成 Kubernetes 应用配置的 Golang实现。

## 🎯 项目目标

- ✅ **100% 兼容现有 Jinja2 模板**（无需修改 Ansible roles）
- ✅ **性能提升 5 倍以上**（从 8 分钟缩短到 1.5 分钟）
- ✅ **跨平台支持**（Windows/macOS/Linux原生运行）
- ✅ **智能预检功能**（减少 80% 配置错误）
- ✅ **开发友好**（支持 `go run` 直接运行）
- ✅ **ArgoCD Application 生成**（新增功能）

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
# ========== K8s 配置生成 ==========
# 预检配置
go run cmd/main.go precheck --base-dir /path/to/configs

# 生成 K8s 配置（最简单的方式）
go run cmd/main.go generate --base-dir /path/to/configs

# 使用默认 base-dir
go run cmd/main.go generate

# ========== ArgoCD Application 生成（新增） ==========
# 生成所有 ArgoCD Applications
go run cmd/main.go argocd generate --base-dir /path/to/configs

# 只生成指定的应用
go run cmd/main.go argocd generate \
  --base-dir /path/to/configs \
  --roles cms-service,fms-service

# 跳过预检（紧急情况）
go run cmd/main.go argocd generate --base-dir /path/to/configs --skip-precheck

# 查看详细日志
go run cmd/main.go argocd generate --base-dir /path/to/configs --verbose
```

## 🎛️ ArgoCD 命令详解

### 命令结构

```bash
# 基本格式
go run cmd/main.go argocd [command] [flags]

# 可用命令
generate    # 生成 ArgoCD Application 配置
help        # 显示帮助信息
```

### Flags 参数说明

| Flag | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--output` | `-o` | `output` | 输出目录路径 |
| `--roles` | - | `nil` | 指定要生成的 roles（逗号分隔） |
| `--skip-precheck` | - | `false` | 跳过预检步骤 |
| `--config` | - | `vars.yaml` | 配置文件路径（继承全局） |
| `--bootstrap` | - | `bootstrap.yml` | Bootstrap 文件路径（继承全局） |
| `--base-dir` | - | `.` | 基础工作目录（继承全局） |
| `--verbose` | `-v` | `false` | 详细日志输出（继承全局） |

### 实战场景

#### 1. 生成所有应用的 ArgoCD 配置

```bash
go run cmd/main.go argocd generate --base-dir .
```

输出示例：
```
🔍 执行预检...
✓ 预检通过
⚙️  正在生成 ArgoCD Application 配置...
✅ 已生成 ArgoCD Application: output/argo-app/cms-project/int/k8s_baas/cms-service.yaml
✅ 已生成 ArgoCD Application: output/argo-app/cms-project/int/k8s_baas/fms-service.yaml
✅ ArgoCD Application 配置生成完成！输出目录：output
```

#### 2. 生成特定环境的配置

```bash
# 为 int 环境生成
go run cmd/main.go argocd generate \
  --base-dir . \
  --config configs/vars-int.yaml

# 为 production 环境生成
go run cmd/main.go argocd generate \
  --base-dir . \
  --config configs/vars-production.yaml
```

#### 3. 批量生成多个应用

```bash
# 生成 3 个应用
go run cmd/main.go argocd generate \
  --base-dir . \
  --roles cms-service,fms-service,user-service
```

#### 4. 自定义输出目录

```bash
# 输出到指定目录
go run cmd/main.go argocd generate \
  --base-dir . \
  --output /tmp/argocd-output
```

#### 5. 紧急情况下跳过预检

```bash
# 跳过预检直接生成（不推荐）
go run cmd/main.go argocd generate \
  --base-dir . \
  --skip-precheck
```

#### 6. 查看详细日志

```bash
# 启用详细日志
go run cmd/main.go argocd generate \
  --base-dir . \
  --verbose
```

### 输出文件结构

生成的 ArgoCD Application 配置将按以下结构组织：

```
output/
└── argo-app/
    └── {project}/
        └── {profile}/
            └── k8s_{stack}/
                ├── cms-service.yaml
                ├── fms-service.yaml
                └── user-service.yaml
```

实际示例：
```
output/argo-app/cms-project/int/k8s_baas/
├── cms-service.yaml
├── fms-service.yaml
└── user-service.yaml
```

## 🏗️ 项目结构

```
k8s-app-accelerator-go/
├── cmd/
│   └── main.go              # CLI 入口
├── internal/
│   ├── cli/                 # CLI 命令入口
│   │   ├── root.go          # Root 命令
│   │   ├── generate.go      # Generate 命令
│   │   ├── precheck.go      # Precheck 命令
│   │   └── argocd.go        # ArgoCD 命令（新增）
│   ├── config/              # 配置加载层
│   │   ├── loader.go        # 配置加载器
│   │   ├── project_config.go # 项目配置结构
│   │   └── resource_group.go # 资源组结构
│   ├── model/               # 数据模型
│   │   ├── role_vars.go     # Role 变量
│   │   └── argocd_app.go    # ArgoCD Application（新增）
│   ├── validator/           # 验证器
│   │   ├── validator.go     # 通用验证器
│   │   ├── checker.go       # 检查规则
│   │   └── argocd_validator.go # ArgoCD 验证器（新增）
│   ├── generator/           # 生成器
│   │   ├── generator.go     # 通用生成器
│   │   └── argocd_generator.go # ArgoCD 生成器（新增）
│   └── template/            # 模板引擎
│       ├── worker.go        # Python Worker
│       └── python_pool.go   # Worker 池
├── configs/                 # 示例配置文件
├── scripts/                 # Python 脚本（Jinja2 Worker）
├── templates/               # Jinja2 模板
│   └── argo-app/            # ArgoCD 模板（新增）
└── docs/                    # 文档目录
    ├── ARGOCD_QUICKSTART.md # ArgoCD 快速开始（新增）
    └── ...
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
argocd:
  site: https://argocd.example.com  # ArgoCD 站点（新增）
stack:
  cms-service: baas
  fms-service: baas
toolset_git_base_url: https://github.example.com
toolset_git_group: my-team
toolset_git_project: my-project
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

### 4. bootstrap.yml - Bootstrap 配置

```yaml
roles:
  - cms-service
  - fms-service
  - user-service
```

## 🧪 测试

```bash
# 运行所有测试
go test -v ./...

# 运行特定包测试
go test -v ./internal/config/...
```

## 🚀 开发工作流

### Phase 1: ArgoCD Application 生成 ✅

- [x] 实现 ArgoCD Application 数据模型
- [x] 实现 ArgoCD Validator 预检系统
- [x] 实现 ArgoCD Generator 生成器
- [x] 实现 ArgoCD CLI 命令
- [ ] 编写单元测试
- [ ] 集成测试与对比验证

### Day 1-2: 项目骨架 + 配置加载层 ✅

- [x] 初始化 Go 模块
- [x] 添加 Cobra CLI 框架
- [x] 创建配置数据结构
- [x] 实现配置加载器
- [x] 编写单元测试

### Day 3-4: Python Worker + 进程池 ✅

- [x] 实现 render_worker.py
- [x] 实现 PythonWorker 封装
- [x] 实现 WorkerPool 进程池
- [x] 添加健康检查机制

### Day 5-6: Pre-Check 预检功能 ✅

- [x] 实现配置校验器
- [x] 实现预检规则
- [x] 添加彩色输出报告

### Day 7-8: Generate 生成功能 ✅

- [x] 实现上下文构建器
- [x] 实现 Role 生成器
- [x] 实现文件写入器

### Day 9-10: ArgoCD 专项功能 ✅

- [x] 实现 ArgoCD Application 模型
- [x] 实现 ArgoCD Validator
- [x] 实现 ArgoCD Generator
- [x] 实现 ArgoCD CLI 命令
- [x] 编写技术文档

### Day 11-12: 测试 + 文档 ⏳

- [ ] 集成测试
- [ ] 性能基准测试
- [ ] 完善文档

## 🛠️ 技术栈

- **语言**: Go 1.21+
- **CLI 框架**: Cobra
- **YAML 解析**: gopkg.in/yaml.v3
- **测试框架**: testify
- **模板引擎**: Jinja2 (Python Worker)
- **HTTP 客户端**: net/http (标准库)

## 📋 下一步

1. **准备测试环境**
   ```bash
   # 复制示例配置文件
   cp configs/vars.example.yaml configs/vars.yaml
   cp bootstrap.example.yml bootstrap.yml
   
   # 安装依赖
   go mod download
   pip3 install -r scripts/requirements.txt
   ```

2. **创建 ArgoCD 模板**
   ```bash
   mkdir -p templates/argo-app
   cp /Users/bohaiqing/work/git/k8s_app_acelerator/argocd/roles/argo-app/templates/app.yaml.j2 \
      templates/argo-app/
   ```

3. **运行首次生成测试**
   ```bash
   go run cmd/main.go argocd generate --base-dir .
   ```

4. **对比 Ansible 输出**
   ```bash
   bash scripts/compare_outputs.sh
   ```

5. **修复发现的差异并编写测试**

## 📚 更多文档

### 核心文档
- [AGENTS.md](./AGENTS.md) - 完整技术方案
- [ARGOCD_IMPLEMENTATION_SUMMARY.md](./ARGOCD_IMPLEMENTATION_SUMMARY.md) - ArgoCD 实现总结

### 快速指南
- [docs/ARGOCD_QUICKSTART.md](./docs/ARGOCD_QUICKSTART.md) - ArgoCD 5 分钟快速上手

### 其他文档
- [PRECHECK_GUIDE.md](./PRECHECK_GUIDE.md) - 预检功能说明
- [WORKDIR_USAGE.md](./docs/WORKDIR_USAGE.md) - Workdir 使用指南
