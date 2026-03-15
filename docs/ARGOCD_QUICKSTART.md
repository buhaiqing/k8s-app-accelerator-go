# ArgoCD 快速开始指南

## 🚀 5 分钟快速上手

### 1. 准备环境

```bash
# 安装 Go 依赖
go mod download

# 安装 Python 依赖
pip3 install Jinja2 PyYAML
```

### 2. 准备配置文件

#### vars.yaml

```yaml
project: cms-project
argocd:
  site: https://argocd.example.com
stack:
  cms-service: baas
  fms-service: baas
toolset_git_base_url: https://github.example.com
toolset_git_group: my-team
toolset_git_project: my-project
```

#### bootstrap.yml

```yaml
roles:
  - cms-service
  - fms-service
```

### 3. 运行生成

```bash
# 进入项目目录
cd /Users/bohaiqing/opensource/git/k8s-app-accelerator-go

# 生成所有 ArgoCD Applications
go run . argocd generate \
  --base-dir . \
  --config configs/vars.yaml \
  --bootstrap bootstrap.yml

# 查看输出
tree output/argo-app
```

### 4. 验证结果

```bash
# 应该看到类似输出：
output/argo-app/
└── cms-project
    └── int
        └── k8s_baas
            ├── cms-service.yaml
            └── fms-service.yaml
```

---

## 💡 常用命令

### 生成所有应用

```bash
go run . argocd generate --base-dir .
```

### 只生成指定的应用

```bash
# 单个应用
go run . argocd generate \
  --base-dir . \
  --roles cms-service

# 多个应用
go run . argocd generate \
  --base-dir . \
  --roles cms-service,fms-service
```

### 跳过预检

```bash
go run . argocd generate \
  --base-dir . \
  --skip-precheck
```

### 查看详细日志

```bash
go run . argocd generate \
  --base-dir . \
  --verbose
```

---

## 🔍 故障排查

### 问题 1: Python Worker 启动失败

**错误**:
```
创建 worker 池失败：exec: "python3": executable file not found in $PATH
```

**解决**:
```bash
# 检查 Python 是否安装
which python3

# 如果未安装，使用 Homebrew 安装（macOS）
brew install python@3.12
```

### 问题 2: 模板文件未找到

**错误**:
```
渲染模板失败：template app.yaml.j2 does not exist
```

**解决**:
```bash
# 创建模板目录
mkdir -p templates/argo-app

# 复制模板文件
cp /Users/bohaiqing/work/git/k8s_app_acelerator/argocd/roles/argo-app/templates/app.yaml.j2 \
   templates/argo-app/
```

### 问题 3: Stack 未定义

**错误**:
```
预检发现 1 个错误
  ✖ stack.cms-service
    问题：应用 cms-service 未定义 Stack
```

**解决**:
在 vars.yaml 中添加 stack 配置：
```yaml
stack:
  cms-service: baas
  fms-service: baas
```

---

## 📁 目录结构

```
k8s-app-accelerator-go/
├── cmd/
│   └── main.go              # CLI 入口
├── internal/
│   ├── model/
│   │   └── argocd_app.go    # ArgoCD Application 模型
│   ├── validator/
│   │   └── argocd_validator.go  # 验证器
│   ├── generator/
│   │   └── argocd_generator.go  # 生成器
│   └── cli/
│       └── argocd.go        # CLI 命令
├── templates/
│   └── argo-app/
│       └── app.yaml.j2      # ArgoCD 模板
├── configs/
│   ├── vars.yaml            # 项目配置
│   └── bootstrap.yml        # Bootstrap 配置
└── output/
    └── argo-app/            # 生成的文件
```

---

## 🎯 下一步

1. **阅读完整文档**: [`ARGOCD_IMPLEMENTATION_SUMMARY.md`](file:///Users/bohaiqing/opensource/git/k8s-app-accelerator-go/ARGOCD_IMPLEMENTATION_SUMMARY.md)

2. **查看技术方案**: [`AGENTS.md`](file:///Users/bohaiqing/opensource/git/k8s-app-accelerator-go/AGENTS.md)

3. **运行对比测试**: 
   ```bash
   bash scripts/compare_harness.sh
   ```

---

**最后更新**: 2026-03-14  
**状态**: ✅ 编码完成
