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
# Go (1.21+) - macOS
brew install go

# Go (1.21+) - Linux
# 访问 https://go.dev/dl/ 下载安装

# Python 依赖
pip3 install -r scripts/requirements.txt
```

### 运行

```bash
# 直接运行（推荐）
go run . --help

# 或编译后运行
go build -o k8s-gen && ./k8s-gen --help
```

### 使用示例

```bash
# ========== K8s 配置生成 ==========
# 预检配置
go run . precheck --base-dir /path/to/configs

# 生成 K8s 配置（最简单的方式）
go run . generate --base-dir /path/to/configs

# 使用默认 base-dir
go run . generate

# ========== ArgoCD Application 生成 ==========
# 生成所有 ArgoCD Applications
go run . argocd generate --base-dir /path/to/configs

# 只生成指定的应用
go run . argocd generate \
  --base-dir /path/to/configs \
  --roles cms-service,fms-service

# 跳过预检（紧急情况）
go run . argocd generate --base-dir /path/to/configs --skip-precheck

# 查看详细日志
go run . argocd generate --base-dir /path/to/configs --verbose

# ========== Jenkins Jobs 生成 ==========
# 生成 Jenkins Jobs 配置
go run . jenkins generate \
  --base-dir /path/to/configs \
  --output output/jenkins

# ========== CMDB SQL 生成 ==========
# 生成 CMDB 初始化 SQL
go run . cmdb \
  --base-dir /path/to/configs \
  --output output/cmdb

# ========== GitLab Cfg 生成 ==========
# 生成 GitLab 项目配置
go run . gitlab-cfg generate \
  --base-dir /path/to/configs \
  --output output/gitlab-cfg
```

## 🎛️ ArgoCD 命令详解

### 命令结构

```bash
# 基本格式
go run . argocd [command] [flags]

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
go run . argocd generate --base-dir .
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
go run . argocd generate \
  --base-dir . \
  --config configs/vars-int.yaml

# 为 production 环境生成
go run . argocd generate \
  --base-dir . \
  --config configs/vars-production.yaml
```

#### 3. 批量生成多个应用

```bash
# 生成 3 个应用
go run . argocd generate \
  --base-dir . \
  --roles cms-service,fms-service,user-service
```

#### 4. 自定义输出目录

```bash
# 输出到指定目录
go run . argocd generate \
  --base-dir . \
  --output /tmp/argocd-output
```

#### 5. 紧急情况下跳过预检

```bash
# 跳过预检直接生成（不推荐）
go run . argocd generate \
  --base-dir . \
  --skip-precheck
```

#### 6. 查看详细日志

```bash
# 启用详细日志
go run . argocd generate \
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
go run . jenkins [command] [flags]

# 可用命令
generate    # 生成 Jenkins Jobs 配置
help        # 显示帮助信息
```

### 使用示例

```bash
# 生成所有 Jenkins Jobs
go run . jenkins generate --base-dir .

# 自定义输出目录
go run . jenkins generate \
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
go run . cmdb [flags]

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
go run . cmdb \
  --base-dir . \
  --output output/cmdb-sql

# 指定配置文件
go run . cmdb \
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
go run . gitlab-cfg [command] [flags]

# 可用命令
generate    # 生成 GitLab 项目配置
help        # 显示帮助信息
```

### 使用示例

```bash
# 生成 GitLab 项目配置
go run . gitlab-cfg generate \
  --base-dir . \
  --output output/gitlab-configs

# 生成特定项目的配置
go run . gitlab-cfg generate \
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

### 1. compare_harness.sh - GitLab Cfg 输出对比脚本

对比 Go 和 Ansible 生成的 GitLab Cfg 配置文件。

```bash
bash scripts/compare_harness.sh
```

### 2. compare_cmdb_outputs.sh - CMDB SQL 输出对比脚本

对比 Go 和 Ansible 生成的 CMDB SQL 脚本。

```bash
# 使用默认输出目录
bash scripts/compare_cmdb_outputs.sh

# 指定 Go 输出目录
GO_OUTPUT_DIR=./output/cmdb-test2 bash scripts/compare_cmdb_outputs.sh
```

### 3. render_worker.py - Jinja2 渲染 Worker

Go 程序内部调用，无需手动执行。

```bash
python3 scripts/render_worker.py --worker-mode
```

## 📋 快速命令参考

```bash
# 安装依赖
pip3 install -r scripts/requirements.txt

# 运行 Go 版本生成
go run . generate --base-dir /path/to/configs
go run . argocd generate --base-dir /path/to/configs
go run . cmdb --base-dir /path/to/configs

# 对比输出
bash scripts/compare_harness.sh
bash scripts/compare_cmdb_outputs.sh
```

## 🏗️ 项目结构

```
k8s-app-accelerator-go/
├── cmd/main.go              # CLI 入口
├── internal/
│   ├── cli/                 # CLI 命令
│   ├── config/              # 配置加载
│   ├── generator/           # 生成器
│   └── template/           # 模板引擎
├── scripts/                 # 辅助脚本
└── docs/                    # 文档
```

## 📝 配置

```yaml
# vars.yaml
project: my-project
profiles: [int, uat, production]
stack:
  cms-service: baas
```
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

## 🛠️ 技术栈

- Go 1.21+ | Cobra | Jinja2

## 📚 文档

- [docs/](./docs/) - 详细技术文档
