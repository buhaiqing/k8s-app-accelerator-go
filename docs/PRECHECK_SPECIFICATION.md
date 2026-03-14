# Pre-Check 预检规范

**前置条件**: 
- ✅ 已阅读 [CLI_REFERENCE.md](./CLI_REFERENCE.md)
- ✅ 了解配置文件结构
- ✅ 运行过 precheck 命令

**目标读者**: 运维人员、开发者  
**最后更新**: 2025-03-14

---

## 📋 检查项总览

| 类别 | 检查项数量 | 错误级别 | 说明 |
|------|----------|---------|------|
| A. 配置文件格式 | 7 | error/warning | vars.yaml, bootstrap.yml 基础校验 |
| B. Resources 完整性 | 4 | error | 数据库、Redis 等资源定义 |
| C. Mapping 一致性 | 3 | warning | 应用映射关系验证 |
| D. Role Vars 完整性 | 8 | error/warning | Role 配置完整性 |
| E. ArgoCD 专项 | 6 | error | ArgoCD Application 特殊检查 |
| F. 模板文件 | 3 | error | Jinja2 模板存在性 |

**总计**: 31 个检查项

---

## A. 配置文件格式检查

### A1. 项目名称不能为空

```yaml
# ❌ 错误示例
project: ""

# ✅ 正确示例
project: dly
```

**检查规则**: `len(project) > 0`  
**错误级别**: error  
**错误提示**: `项目名称不能为空`

---

### A2. 项目名称格式

```yaml
# ❌ 错误示例
project: "DLY-Project"  # 包含大写字母和特殊字符

# ✅ 正确示例
project: dly  # 只能包含小写字母和数字
```

**检查规则**: `^[a-z0-9]+$`  
**错误级别**: error  
**错误提示**: `项目名称只能包含小写字母和数字`

---

### A3. 至少定义一个环境 (profile)

```yaml
# ❌ 错误示例
profiles: []

# ✅ 正确示例
profiles:
  - int
  - production
```

**检查规则**: `len(profiles) >= 1`  
**错误级别**: error  
**错误提示**: `至少定义一个环境 (profile)`

---

### A4. profile 名称规范性

```yaml
# ⚠️ 警告示例
profiles:
  - dev      # 非标准名称
  - prod     # 非标准名称

# ✅ 推荐示例
profiles:
  - int      # 集成测试
  - uat      # 用户验收测试
  - production # 生产环境
```

**检查规则**: profile in [`int`, `uat`, `production`]  
**错误级别**: warning  
**警告提示**: `推荐使用标准 profile 名称：int, uat, production`

---

### A5. Apollo Token 格式验证

```yaml
# ❌ 错误示例
apollo:
  token: "invalid-token"

# ✅ 正确示例
apollo:
  token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."  # JWT 格式
```

**检查规则**: `strings.HasPrefix(token, "eyJ")`  
**错误级别**: warning  
**警告提示**: `Apollo Token 应该是 JWT 格式`

---

### A6. ArgoCD 地址配置检查

```yaml
# ❌ 错误示例
argocd:
  addr: ""  # 空值

# ✅ 正确示例
argocd:
  addr: "https://argocd.example.com"
```

**检查规则**: `addr != "" && strings.HasPrefix(addr, "http")`  
**错误级别**: error  
**错误提示**: `ArgoCD 地址必须配置且以 http(s) 开头`

---

### A7. Git 仓库 URL 格式验证

```yaml
# ❌ 错误示例
toolset_git_base_url: "github.com/xxx"  # 缺少协议

# ✅ 正确示例
toolset_git_base_url: "https://github.com/xxx"
```

**检查规则**: `strings.HasPrefix(url, "http") || strings.HasPrefix(url, "git@")`  
**错误级别**: error  
**错误提示**: `Git 仓库 URL 必须以 http(s) 或 git@开头`

---

## B. Resources 完整性检查

### B1. 默认 RDS 连接地址必须配置

```yaml
# ❌ 错误示例
rds: []

# ✅ 正确示例
rds:
  - name: default
    host: rm-xxx.mysql.rds.aliyuncs.com
    port: 3306
```

**检查规则**: `exists(rds[?name=='default'])`  
**错误级别**: error  
**错误提示**: `必须配置名为 default 的 RDS 连接`

---

### B2. 数据库端口范围

```yaml
# ❌ 错误示例
rds:
  - port: 70000  # 超出范围

# ✅ 正确示例
rds:
  - port: 3306  # 有效端口
```

**检查规则**: `port >= 1 && port <= 65535`  
**错误级别**: error  
**错误提示**: `数据库端口必须在 1-65535 范围内`

---

### B3. 密码强度检查

```yaml
# ⚠️ 警告示例
password: "123456"  # 弱密码

# ✅ 推荐示例
password: "Hd@2025#SecurePass123"  # 强密码
```

**检查规则**: 
- 长度 >= 12
- 包含大小写字母
- 包含数字
- 包含特殊字符

**错误级别**: warning  
**警告提示**: `建议使用强密码（大小写 + 数字 + 特殊字符，长度≥12）`

---

### B4. Redis 端口安全提示

```yaml
# ⚠️ 警告示例
redis:
  - port: 6379  # 默认端口

# ✅ 推荐示例
redis:
  - port: 16379  # 自定义端口
```

**检查规则**: `port == 6379`  
**错误级别**: warning  
**警告提示**: `建议使用非标准 Redis 端口以提高安全性`

---

## C. Mapping 一致性检查

### C1. 每个 role 在 mapping 中有定义

```yaml
# ❌ 错误示例
# bootstrap.yml 定义了 gateway-service
# mapping.yaml 没有 gateway-service

# ✅ 正确示例
mappings:
  gateway-service:
    product: baas
```

**检查规则**: `forall(role in bootstrap) exists(mapping[role])`  
**错误级别**: error  
**错误提示**: `Role {role} 在 mapping.yaml 中没有定义`

---

### C2. product 值不能为空

```yaml
# ❌ 错误示例
gateway-service:
  product: ""

# ✅ 正确示例
gateway-service:
  product: baas
```

**检查规则**: `product != ""`  
**错误级别**: error  
**错误提示**: `product 值不能为空`

---

### C3. product 格式规范

```yaml
# ⚠️ 警告示例
product: "BaaS"  # 大写

# ✅ 正确示例
product: baas  # 小写
```

**检查规则**: `product =~ ^[a-z_]+$`  
**错误级别**: warning  
**警告提示**: `product 应该使用小写字母和下划线`

---

## D. Role Vars 完整性检查

### D1. app 字段必须定义

```yaml
# ❌ 错误示例
- DNET_PRODUCT: baas

# ✅ 正确示例
- app: gateway-service
  DNET_PRODUCT: baas
```

**检查规则**: `app != ""`  
**错误级别**: error  
**错误提示**: `app 字段必须定义`

---

### D2. DNET_PRODUCT 必须定义

```yaml
# ❌ 错误示例
- app: gateway-service

# ✅ 正确示例
- app: gateway-service
  DNET_PRODUCT: baas
```

**检查规则**: `DNET_PRODUCT != ""`  
**错误级别**: error  
**错误提示**: `DNET_PRODUCT 字段必须定义`

---

### D3. _type 只能是 backend 或 frontend

```yaml
# ❌ 错误示例
_type: service

# ✅ 正确示例
_type: backend  # 或 frontend
```

**检查规则**: `_type in ['backend', 'frontend']`  
**错误级别**: error  
**错误提示**: `_type 只能是 backend 或 frontend`

---

### D4. 前端组件不应启用 enable_rdb

```yaml
# ❌ 错误示例
_type: frontend
enable_rdb: true

# ✅ 正确示例
_type: frontend
enable_rdb: false
```

**检查规则**: `if _type == 'frontend' then enable_rdb == false`  
**错误级别**: warning  
**警告提示**: `前端组件不应启用 RDB`

---

### D5. CPU limits >= requests

```yaml
# ❌ 错误示例
cpu_requests: "1000m"
cpu_limits: "500m"  # limits < requests

# ✅ 正确示例
cpu_requests: "500m"
cpu_limits: "1000m"
```

**检查规则**: `parse_cpu(limits) >= parse_cpu(requests)`  
**错误级别**: error  
**错误提示**: `CPU limits 必须 >= requests`

---

### D6. Memory limits >= requests

```yaml
# ❌ 错误示例
memory_requests: "2Gi"
memory_limits: "1Gi"  # limits < requests

# ✅ 正确示例
memory_requests: "1Gi"
memory_limits: "2Gi"
```

**检查规则**: `parse_mem(limits) >= parse_mem(requests)`  
**错误级别**: error  
**错误提示**: `Memory limits 必须 >= requests`

---

### D7. 内存请求合理性检查

```yaml
# ⚠️ 警告示例
memory_requests: "16Gi"  # 过大

# ✅ 推荐示例
memory_requests: "2Gi"  # 合理值
```

**检查规则**: `parse_mem(requests) > 8Gi`  
**错误级别**: warning  
**警告提示**: `内存请求超过 8GB，请确认是否合理`

---

## E. ArgoCD Application 专项检查

### E1. Stack 映射存在性检查

```yaml
# ❌ 错误示例
# stack.yaml 没有定义 gateway-service 的 stack

# ✅ 正确示例
stack:
  gateway-service: zt4d
```

**检查规则**: `forall(app) exists(stack[app])`  
**错误级别**: error  
**错误提示**: `应用 {app} 的 stack 未定义`

---

### E2. Git 分支名称规范性

```yaml
# ⚠️ 警告示例
git_branch: "feature/new-feature"  # 包含斜杠

# ✅ 推荐示例
git_branch: "k8s_mas"  # 简单命名
```

**检查规则**: `!strings.Contains(branch, "/")`  
**错误级别**: warning  
**警告提示**: `Git 分支名称不应包含斜杠`

---

### E3. Kustomize 版本兼容性

```yaml
# ⚠️ 警告示例
kustomize_version: "3.0.0"  # 过旧

# ✅ 推荐示例
kustomize_version: "5.0.0"  # 最新稳定版
```

**检查规则**: `version >= "4.0.0"`  
**错误级别**: warning  
**警告提示**: `建议使用 Kustomize 4.0+`

---

### E4. Destination namespace 有效性

```yaml
# ❌ 错误示例
namespace: ""  # 空值

# ✅ 正确示例
namespace: baas  # 有效的 namespace
```

**检查规则**: `namespace != ""`  
**错误级别**: error  
**错误提示**: `Destination namespace 必须配置`

---

### E5. SyncPolicy 配置正确性

```yaml
# ⚠️ 警告示例
syncPolicy: {}  # 空的 sync policy

# ✅ 推荐示例
syncPolicy:
  automated:
    prune: true
    selfHeal: true
```

**检查规则**: `syncPolicy != {}`  
**错误级别**: warning  
**警告提示**: `建议配置 syncPolicy`

---

### E6. Finalizers 配置必要性

```yaml
# ⚠️ 警告示例
finalizers: []  # 没有 finalizers

# ✅ 推荐示例
finalizers:
  - resources-finalizer.argocd.argoproj.io
```

**检查规则**: `len(finalizers) > 0`  
**错误级别**: warning  
**警告提示**: `建议配置 finalizers 以确保资源清理顺序`

---

## F. 模板文件存在性检查

### F1. app.yaml.j2 (ArgoCD) 存在

```bash
# ❌ 错误示例
templates/argo-app/app.yaml.j2  # 文件不存在

# ✅ 正确示例
templates/argo-app/app.yaml.j2  # 文件存在
```

**检查规则**: `file_exists(templates/argo-app/app.yaml.j2)`  
**错误级别**: error  
**错误提示**: `ArgoCD 模板文件不存在`

---

### F2. job.j2 (Jenkins) 存在

```bash
# ❌ 错误示例
templates/jenkins-jobs/job.j2  # 文件不存在

# ✅ 正确示例
templates/jenkins-jobs/job.j2  # 文件存在
```

**检查规则**: `file_exists templates/jenkins-jobs/job.j2)`  
**错误级别**: error  
**错误提示**: `Jenkins 模板文件不存在`

---

### F3. sql.j2 (CMDB) 存在

```bash
# ❌ 错误示例
templates/cmdb/sql.j2  # 文件不存在

# ✅ 正确示例
templates/cmdb/sql.j2  # 文件存在
```

**检查规则**: `file_exists(templates/cmdb/sql.j2)`  
**错误级别**: error  
**错误提示**: `CMDB 模板文件不存在`

---

## 📊 检查结果输出

### 成功示例

```
✓ 配置文件格式检查通过 (7/7)
✓ Resources 完整性检查通过 (4/4)
✓ Mapping 一致性检查通过 (3/3)
✓ Role Vars 完整性检查通过 (8/8)
✓ ArgoCD 专项检查通过 (6/6)
✓ 模板文件存在性检查通过 (3/3)

✅ 预检通过，可以安全生成配置
```

### 失败示例

```
✖ 配置文件格式检查发现 2 个问题:
  ✖ 项目名称不能为空
    问题：project 字段为空字符串
    建议：在 vars.yaml 中配置有效的项目名称
  
  ✖ ArgoCD 地址配置缺失
    问题：argocd.addr 未配置
    建议：配置 ArgoCD 服务器地址，如 https://argocd.example.com

❌ 预检失败，请修复上述问题后重试
```

---

## 🔧 使用建议

### 1. 开发阶段频繁使用

```bash
# 每次修改配置后运行
go run cmd/main.go precheck --base-dir configs
```

### 2. CI/CD 集成

```yaml
# .github/workflows/ci.yml
- name: Pre-check Configuration
  run: go run cmd/main.go precheck --base-dir configs
```

### 3. 批量预检多个环境

```bash
for env in int production; do
  echo "Checking $env..."
  go run cmd/main.go precheck \
    --base-dir configs/$env
done
```

---

## 📚 相关文档

- [CLI_REFERENCE.md](./CLI_REFERENCE.md) - precheck 命令参考
- [ARCHITECTURE_DEEP_DIVE.md](./ARCHITECTURE_DEEP_DIVE.md) - Validator 实现细节
- [BEST_PRACTICES.md](./BEST_PRACTICES.md) - 配置管理最佳实践

---

**最后更新**: 2025-03-14  
**维护者**: K8s App Accelerator Team
