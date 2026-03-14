# CLI 命令参考手册

**前置条件**: 无  
**目标读者**: 所有用户  
**最后更新**: 2025-03-14

---

## 📋 命令总览

```bash
k8s-gen <command> [flags]
```

### 可用命令

| 命令 | 功能 | 使用场景 |
|------|------|---------|
| `generate` | 生成 K8s 配置 | 复用 Ansible roles 生成 Kubernetes 配置 |
| `argocd generate` | 生成 ArgoCD 配置 | 批量生成 ArgoCD Application CRD |
| `jenkins generate` | 生成 Jenkins 配置 | 批量生成 Jenkins Job 配置 |
| `cmdb` | 生成 CMDB SQL | 生成数据库初始化脚本 |
| `precheck` | 预检配置 | 验证配置文件完整性和正确性 |
| `version` | 显示版本 | 查看工具版本信息 |

---

## 🎯 全局 Flags

| Flag | 简写 | 说明 | 默认值 | 必填 |
|------|------|------|--------|------|
| `--base-dir` | `-b` | 基础目录路径（读取 configs/*） | 当前目录 | ❌ |
| `--config` | | 配置文件路径 | `configs/vars.yaml` | ❌ |
| `--bootstrap` | | Bootstrap 文件路径 | `bootstrap.yml` | ❌ |
| `--resources` | | 资源文件路径 | `configs/resources.yaml` | ❌ |
| `--mapping` | | Mapping 文件路径 | `configs/mapping.yaml` | ❌ |
| `--output` | `-o` | 输出目录 | `output` | ❌ |
| `--roles` | | 指定要生成的 roles | 全部 | ❌ |
| `--skip-precheck` | | 跳过预检 | false | ❌ |
| `--verbose` | `-v` | 详细日志输出 | false | ❌ |
| `--workdir` | `-w` | 工作目录 | 当前目录 | ❌ |

---

## 📖 命令详解

### 1. generate - 生成 K8s 配置

**功能**: 复用现有 Ansible roles 生成 Kubernetes 配置

#### 基本用法

```bash
# 标准方式（推荐）
go run cmd/main.go generate --base-dir /path/to/configs

# 指定工作目录
go run cmd/main.go generate \
  --workdir /path/to/project \
  --base-dir configs
```

#### 自定义配置文件名

```bash
go run cmd/main.go generate \
  --base-dir /path/to/configs \
  --bootstrap bootstrap-test.yml \
  --vars vars-test.yaml \
  --resources resources-test.yaml \
  --mapping mapping-test.yaml
```

#### 生成指定的 roles

```bash
go run cmd/main.go generate \
  --base-dir configs \
  --roles cms-service,fms-service
```

#### 输出结构

```
output/
└── {project}/
    └── {profile}/
        └── k8s_{stack}/
            ├── {app1}.yaml
            └── {app2}.yaml
```

---

### 2. argocd generate - 生成 ArgoCD 配置

**功能**: 批量生成 ArgoCD Application CRD 配置文件

#### 基本用法

```bash
# 批量生成所有应用
go run cmd/main.go argocd generate --base-dir /path/to/configs

# 指定输出目录
go run cmd/main.go argocd generate \
  --base-dir configs \
  -o output/argo-app
```

#### 生成指定的应用

```bash
go run cmd/main.go argocd generate \
  --base-dir configs \
  --roles gateway-service,config-service
```

#### 跳过预检

```bash
go run cmd/main.go argocd generate \
  --base-dir configs \
  --skip-precheck
```

#### 输出结构

```
output/argo-app/
└── {project}/
    └── {profile}/
        └── k8s_{stack}/
            ├── {app1}.yaml
            └── {app2}.yaml
```

#### 配置文件要求

```yaml
# configs/vars.yaml
common: &common
  project: dly
  profiles:
    - int
    - production
  stack:
    gateway-service: zt4d
    config-service: zt4d

# bootstrap.yml
roles:
  - gateway-service
  - config-service
profile: int  # 可选，覆盖所有 role 的 profile
```

---

### 3. jenkins generate - 生成 Jenkins 配置

**功能**: 批量生成 Jenkins Job 配置文件

#### 基本用法

```bash
# 批量生成所有产品
go run cmd/main.go jenkins generate --base-dir /path/to/configs

# 指定输出目录
go run cmd/main.go jenkins generate \
  --base-dir configs \
  -o output/jenkins
```

#### 生成指定的产品

```bash
go run cmd/main.go jenkins generate \
  --base-dir configs \
  --roles baas,mas,cms
```

#### 跳过预检

```bash
go run cmd/main.go jenkins generate \
  --base-dir configs \
  --skip-precheck
```

#### 输出结构

```
output/jenkins/
├── baas/
│   └── project.yml
├── mas/
│   └── project.yml
└── cms/
    └── project.yml
```

#### 配置文件要求

直接复用 Ansible 项目的 `vars.yaml` 文件：

```yaml
# configs/vars.yaml
common: &common
  DNET_PROJECT: zhseczt
  GIT_BASE_URL: https://github-argocd.hd123.com/
  GIT_BASE_GROUP: qianfanops
  output: output
  receivers: x@hd123.com
  env: '测试环境 K8S'
  surfix: Int

data:
  - <<: *common
    DNET_PRODUCT: baas
    product_des: '中台'
  - <<: *common
    DNET_PRODUCT: mas
    product_des: '资料中台'
```

---

### 4. cmdb - 生成 CMDB SQL

**功能**: 生成 CMDB 数据库初始化 SQL 脚本

#### 基本用法

```bash
# 生成 SQL 脚本
go run cmd/main.go cmdb --base-dir /path/to/configs

# 指定输出目录
go run cmd/main.go cmdb \
  --base-dir configs \
  -o output/cmdb
```

#### 自定义配置文件名

```bash
go run cmd/main.go cmdb \
  --base-dir configs \
  --vars vars-prod.yaml \
  --resources resources-prod.yaml
```

#### 输出结构

```
output/cmdb/
├── inittables.sql
└── {profile}.sql
```

#### 配置文件要求

```yaml
# configs/vars.yaml
project: dly
profiles:
  - int
  - production

# configs/resources.yaml
rds:
  - name: default
    host: rm-xxx.mysql.rds.aliyuncs.com
    port: 3306
    user: root
    password: xxx
```

---

### 5. precheck - 预检配置

**功能**: 验证配置文件的完整性和正确性

#### 基本用法

```bash
# 预检配置文件
go run cmd/main.go precheck --base-dir /path/to/configs

# 查看详细日志
go run cmd/main.go precheck \
  --base-dir configs \
  --verbose
```

#### 检查项

详见：[PRECHECK_SPECIFICATION.md](./PRECHECK_SPECIFICATION.md)

#### 输出示例

```
✓ 配置文件格式检查通过
✓ Resources 完整性检查通过
✓ Mapping 一致性检查通过
✓ Role Vars 完整性检查通过
✓ 模板文件存在性检查通过

✅ 预检通过，可以安全生成配置
```

---

## 🔧 常见问题

### Q1: 如何只生成部分应用的配置？

使用 `--roles` flag 指定要生成的 roles：

```bash
go run cmd/main.go generate \
  --base-dir configs \
  --roles app1,app2,app3
```

### Q2: 如何跳过预检？

在紧急情况下可以使用 `--skip-precheck`：

```bash
go run cmd/main.go generate \
  --base-dir configs \
  --skip-precheck
```

⚠️ **注意**: 不推荐跳过预检，可能导致生成失败

### Q3: 如何查看详细日志？

添加 `-v` 或 `--verbose` flag：

```bash
go run cmd/main.go generate \
  --base-dir configs \
  --verbose
```

### Q4: 如何指定不同的环境？

通过 bootstrap.yml 中的 `profile` 字段控制：

```yaml
# bootstrap.yml
profile: production  # 或 int, uat
```

### Q5: 输出目录在哪里？

默认在当前目录的 `output/` 下，可以通过 `-o` 指定：

```bash
go run cmd/main.go generate \
  --base-dir configs \
  -o /tmp/output
```

---

## 📊 最佳实践

### 1. 使用标准目录结构

```
project-root/
├── bootstrap.yml          # Bootstrap 配置
├── configs/
│   ├── vars.yaml         # 项目配置
│   ├── resources.yaml    # 资源定义
│   └── mapping.yaml      # 应用映射
└── output/               # 生成结果
```

### 2. 先预检再生成

```bash
# 步骤 1: 预检
go run cmd/main.go precheck --base-dir configs

# 步骤 2: 生成
go run cmd/main.go generate --base-dir configs
```

### 3. 使用 Makefile 简化操作

```makefile
.PHONY: generate precheck

generate:
	go run cmd/main.go generate --base-dir configs

precheck:
	go run cmd/main.go precheck --base-dir configs
```

### 4. 批量处理多个环境

```bash
# 为不同环境生成配置
for profile in int production; do
  go run cmd/main.go generate \
    --base-dir configs \
    -o output/${profile}
done
```

---

## 📚 相关文档

- [开发指南](./DEVELOPMENT_GUIDE.md) - 环境搭建和快速开始
- [预检规范](./PRECHECK_SPECIFICATION.md) - 详细的检查项说明
- [架构解析](./ARCHITECTURE_DEEP_DIVE.md) - 命令层实现细节

---

**最后更新**: 2025-03-14  
**维护者**: K8s App Accelerator Team
