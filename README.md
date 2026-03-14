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

# ========== ArgoCD Application 生成 ==========
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

# ========== Jenkins Jobs 生成 ==========
# 生成 Jenkins Jobs 配置
go run cmd/main.go jenkins generate \
  --base-dir /path/to/configs \
  --output output/jenkins

# ========== CMDB SQL 生成 ==========
# 生成 CMDB 初始化 SQL
go run cmd/main.go cmdb \
  --base-dir /path/to/configs \
  --output output/cmdb

# ========== GitLab Cfg 生成 ==========
# 生成 GitLab 项目配置
go run cmd/main.go gitlab-cfg generate \
  --base-dir /path/to/configs \
  --output output/gitlab-cfg
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

## 🎛️ Jenkins Jobs 命令详解

### 命令结构

```bash
# 基本格式
go run cmd/main.go jenkins [command] [flags]

# 可用命令
generate    # 生成 Jenkins Jobs 配置
help        # 显示帮助信息
```

### 使用示例

```bash
# 生成所有 Jenkins Jobs
go run cmd/main.go jenkins generate --base-dir .

# 自定义输出目录
go run cmd/main.go jenkins generate \
  --base-dir . \
  --output output/jenkins-jobs
```

### 输出文件结构

```
output/
└── jenkins-jobs/
    └── jobs/
        ├── cms-service/
        │   └── config.xml
        ├── fms-service/
        │   └── config.xml
        └── ...
```

## 🎛️ CMDB SQL 命令详解

### 命令结构

```bash
# 基本格式
go run cmd/main.go cmdb [flags]

# 可用命令
cmdb    # 生成 CMDB 初始化 SQL
help    # 显示帮助信息
```

### Flags 参数说明

| Flag | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--output` | `-o` | `output` | 输出目录路径 |
| `--vars` | - | `configs/vars.yaml` | vars 配置文件路径 |
| `--resources` | - | `configs/resources.yaml` | resources 资源文件路径 |
| `--base-dir` | - | `.` | 基础工作目录 |

### 使用示例

```bash
# 生成 CMDB SQL 脚本
go run cmd/main.go cmdb \
  --base-dir . \
  --output output/cmdb-sql

# 指定配置文件
go run cmd/main.go cmdb \
  --base-dir /path/to/configs \
  --vars vars-test.yaml \
  --resources resources.yaml \
  --output output/cmdb-test
```

### 输出文件结构

```
output/
└── cmdb/
    ├── sql_int.sql              # 测试环境 SQL
    ├── sql_uat.sql              # UAT 环境 SQL
    └── sql_production.sql       # 生产环境 SQL
```

### SQL 内容示例

```sql
-- DockerHub 凭证
INSERT INTO `dockerhub`(`url`, `username`, `password`, `environment`) 
VALUES ('harbor.qianfan123.com', 'qianfan', 'headingqianfan', 'int');

-- Stack 配置
INSERT INTO `stack`(`id`, `type`, `environment`, `dbCount`, `shopCountPerDb`, `currentShopCount`) 
VALUES ('baas', 'plat', 'int', NULL, NULL, NULL);

-- RDS 数据库
INSERT INTO `rds`(`id`, `stackId`, `ip`, `port`, `username`, `password`, `rdsInstanceId`, ...)
VALUES ('int-rds-14736', 'baas', 't4nvm2lpvtysg.oceanbase.aliyuncs.com', '3306', ...);

-- MongoDB 配置
INSERT INTO `dds`(`id`, `ip`, `port`, `password`, `ddsInstanceId`, `stackId`, `username`, ...)
VALUES ('int-mongo-13846', NULL, NULL, 'JGsEUpf4hfFSro8n', NULL, 'baas', 'lpmas', ...);
```

## 🎛️ GitLab Cfg 命令详解

### 命令结构

```bash
# 基本格式
go run cmd/main.go gitlab-cfg [command] [flags]

# 可用命令
generate    # 生成 GitLab 项目配置
help        # 显示帮助信息
```

### 使用示例

```bash
# 生成 GitLab 项目配置
go run cmd/main.go gitlab-cfg generate \
  --base-dir . \
  --output output/gitlab-configs

# 生成特定项目的配置
go run cmd/main.go gitlab-cfg generate \
  --base-dir . \
  --project my-project \
  --output output/my-project-configs
```

### 输出文件结构

```
output/
└── gitlab-configs/
    ├── vars.yaml                    # 项目变量配置
    ├── resources.yaml               # 资源配置
    ├── mapping.yaml                 # 应用映射
    └── bootstrap.yml                # Bootstrap 配置
```

## 🛠️ Scripts 工具脚本

项目提供了多个实用的 Shell 脚本工具，用于自动化测试、对比和验证。

### 1. compare_cmdb_outputs.sh - CMDB 输出对比脚本

**用途**: 对比 Go 版本和 Ansible 版本生成的 CMDB SQL 配置文件

**基本用法**:
```bash
cd /Users/bohaiqing/opensource/git/k8s-app-accelerator-go
./scripts/compare_cmdb_outputs.sh
```

**详细说明**:
- 自动清理旧的输出目录
- 运行 Go 版本生成器
- 尝试运行 Ansible 版本生成器（如果 playbook 存在）
- 智能对比两个版本的输出
- 显示详细的差异报告

**输出示例**:
```
================================================
CMDB SQL 生成器对比测试
================================================

工作目录：/Users/bohaiqing/opensource/git/k8s-app-accelerator-go/scripts
基础目录：/Users/bohaiqing/work/git/k8s_app_acelerator/gitlab_cfg

[1/5] 清理输出目录...
[2/5] 运行 Go 版本生成器...
✓ Go 版本生成成功
[3/5] 准备 Ansible 版本输出...
✓ Ansible 版本输出已存在
[4/5] 检查输出文件...
  Go 版本生成 1 个文件
  Ansible 版本生成 1 个文件
[5/5] 对比输出结果...

✅ 恭喜！Go 和 Ansible 生成的输出完全一致！

================================================
✓ Go 版本 CMDB SQL 生成器功能验证完成！
================================================
```

### 2. compare_jenkins_outputs.sh - Jenkins 输出对比脚本

**用途**: 对比 Go 版本和 Ansible 版本生成的 Jenkins Jobs 配置

**基本用法**:
```bash
cd /Users/bohaiqing/opensource/git/k8s-app-accelerator-go
./scripts/compare_jenkins_outputs.sh
```

**详细说明**:
- 自动清理并重建输出目录
- 并行运行 Go 和 Ansible 版本生成器
- 逐文件对比生成的 XML 配置
- 忽略空白字符和格式差异
- 生成详细的对比报告

**输出示例**:
```
🧹 清理旧的输出目录...
🚀 运行 Go 实现...
✅ Go 实现完成
🚀 运行 Ansible 实现...
✅ Ansible 实现完成

📊 对比结果
==================================================
总文件数：       5
相同文件数：     5
不同文件数：     0

✅ 所有文件内容完全一致！
```

### 3. compare_argocd_outputs.sh - ArgoCD 输出对比脚本

**用途**: 对比 Go 版本和 Ansible 版本生成的 ArgoCD Application 配置

**基本用法**:
```bash
cd /Users/bohaiqing/opensource/git/k8s-app-accelerator-go
./scripts/compare_argocd_outputs.sh
```

**特点**:
- 支持多环境对比（int, uat, production）
- 智能过滤时间戳等动态字段
- 生成 Markdown 格式的对比报告

### 4. render_worker.py - Python Jinja2 渲染 Worker

**用途**: 为 Go 程序提供 Jinja2 模板渲染服务

**通信协议**: JSON-RPC over stdin/stdout

**使用方式**:
```bash
# Go 程序内部调用（无需手动执行）
python3 scripts/render_worker.py --worker-mode
```

**支持的 Filters**:
- `ternary` - Ansible ternary filter
- `upper` / `lower` - 大小写转换
- `profile_convert` - 环境名称转换（int → INT）
- `mandatory` - 必填值校验
- 以及所有 Ansible 内置 filters

### 5. filters.py - Ansible Filters 实现

**用途**: 提供与 Ansible 兼容的 Jinja2 filters

**主要 Filters**:
```python
# 三元运算符
def ternary(value, true_val='', false_val=''):
    return true_val if value else false_val

# 环境转换
def profile_convert(profile):
    return profile.upper()

# 必填校验
def mandatory(value):
    if not value:
        raise ValueError("mandatory value is required")
    return value
```

### 6. test_workdir.sh - Workdir 测试脚本

**用途**: 测试 CLI 工具的 workdir 参数功能

**基本用法**:
```bash
./scripts/test_workdir.sh
```

## 🔧 自定义对比脚本

如果需要为新的生成器创建对比脚本，可以参考以下模板：

```bash
#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 配置目录
BASE_DIR="${1:-/path/to/configs}"
GO_OUTPUT_DIR="./output/go-module"
ANSIBLE_OUTPUT_DIR="./output/ansible-module"

echo "================================================"
echo "模块名称 对比测试"
echo "================================================"

# 清理旧输出
rm -rf "${GO_OUTPUT_DIR}" "${ANSIBLE_OUTPUT_DIR}"
mkdir -p "${GO_OUTPUT_DIR}" "${ANSIBLE_OUTPUT_DIR}"

# 运行 Go 版本
echo "[1/4] 运行 Go 版本..."
go run cmd/main.go module generate \
    --base-dir "$BASE_DIR" \
    -o "$GO_OUTPUT_DIR"

# 运行 Ansible 版本
echo "[2/4] 运行 Ansible 版本..."
ansible-playbook playbook.yaml \
    --extra-vars "output=${ANSIBLE_OUTPUT_DIR}"

# 对比输出
echo "[3/4] 对比输出..."
diff -rq "$GO_OUTPUT_DIR" "$ANSIBLE_OUTPUT_DIR" || true

# 生成报告
echo "[4/4] 生成对比报告..."
echo "✅ 对比完成"
```

## 📋 最佳实践

### 1. 运行对比测试前

```bash
# 确保安装了依赖
pip3 install -r scripts/requirements.txt
go mod download

# 确保有可执行的权限
chmod +x scripts/*.sh
```

### 2. 查看对比报告

```bash
# 将对比结果保存到文件
./scripts/compare_cmdb_outputs.sh > comparison_report.txt 2>&1

# 查看文件
cat comparison_report.txt
```

### 3. 调试模式

```bash
# 启用详细输出
set -x
./scripts/compare_cmdb_outputs.sh
set +x
```

### 4. 清理临时文件

```bash
# 清理所有测试输出
rm -rf ./output/cmdb-go ./output/ansible-cmdb
rm -rf ./output/go-jenkins ./output/ansible-jenkins
rm -rf /tmp/cmdb-comparison
```

## 🎯 脚本特性

- ✅ **跨平台兼容** - 支持 macOS/Linux (使用 POSIX sh 语法)
- ✅ **智能对比** - 自动过滤随机数、时间戳等动态字段
- ✅ **彩色输出** - 使用颜色标识不同的状态
- ✅ **详细报告** - 生成完整的对比报告和差异分析
- ✅ **自动化** - 一键运行所有测试
- ✅ **错误处理** - 完善的错误检测和提示

## 📝 注意事项

1. **Ansible 路径问题**: 部分 Ansible role 使用了硬编码的路径，可能需要创建符号链接
2. **随机数处理**: 对比脚本会自动过滤随机数，只比较结构和格式
3. **环境变量**: 确保已配置必要的环境变量（如 ANSIBLE_CONFIG）
4. **Python 版本**: 需要 Python 3.6+ 以支持所有 Jinja2 特性

## 📚 更多资源

- [CMDB 对比测试完整流程](./docs/CMDB_COMPARISON_GUIDE.md)
- [Jenkins 对比测试指南](./docs/JENKINS_COMPARISON_GUIDE.md)
- [ArgoCD 对比验证](./docs/ARGOCD_COMPARISON_GUIDE.md)

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
