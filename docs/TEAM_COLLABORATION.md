# 团队协作指南

**前置条件**: 
- ✅ 已是项目贡献者
- ✅ 准备提交代码

**目标读者**: 团队成员  
**最后更新**: 2025-03-14

---

## 📝 Git 提交规范

### Commit Message 格式

```
<type>: <subject>

<body>

<footer>
```

### Type 类型

| Type | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat: add ArgoCD Application generator` |
| `fix` | Bug 修复 | `fix: resolve path separator issue` |
| `docs` | 文档更新 | `docs: update CLI reference guide` |
| `style` | 代码格式 | `style: fix indentation` |
| `refactor` | 重构 | `refactor: extract worker pool logic` |
| `perf` | 性能优化 | `perf: improve rendering speed` |
| `test` | 测试 | `test: add unit tests for validator` |
| `chore` | 构建/工具 | `chore: update dependencies` |

### Subject 规范

- 使用祈使句、现在时态
- 首字母小写
- 不以句号结尾
- 长度不超过 50 字符

**示例**:
```bash
# ✅ 正确
feat: add pre-check validation for Redis configuration
fix: resolve path separator issue on Windows

# ❌ 错误
Added pre-check validation  # 过去时
Fixed the bug  # 太模糊
```

### Body 规范（可选）

- 描述变更动机
- 说明设计决策
- 列出 breaking changes

**示例**:
```markdown
feat: optimize worker pool size

Previous implementation used fixed 5 workers, which was
inefficient for small batches.

Now dynamically adjusts based on workload:
- 1-10 apps: 3 workers
- 11-50 apps: 5 workers
- 51+ apps: 10 workers

BREAKING CHANGE: Worker pool initialization signature changed
```

### Footer 规范（可选）

关联 Issue 或 PR：

```markdown
Closes #123
See also #456
```

---

## 🔍 代码审查清单

### 代码质量

- [ ] 代码是否遵循 Go best practices？
- [ ] 是否有适当的错误处理？
- [ ] 是否有不必要的复杂度？
- [ ] 变量和函数命名是否清晰？
- [ ] 函数长度是否合理（< 50 行）？

### 测试覆盖

- [ ] 是否添加了单元测试？
- [ ] 测试是否覆盖了边界情况？
- [ ] 测试是否可重复运行？
- [ ] 测试断言是否明确？

### 文档完整性

- [ ] 公共 API 是否有注释？
- [ ] 复杂逻辑是否有说明？
- [ ] README 是否需要更新？
- [ ] CHANGELOG 是否需要更新？

### 性能影响

- [ ] 是否有性能回归风险？
- [ ] 内存使用是否合理？
- [ ] 是否需要 benchmark 测试？
- [ ] 并发是否安全？

### 安全性

- [ ] 是否有 SQL 注入风险？
- [ ] 是否有敏感信息泄露？
- [ ] 文件权限是否正确？
- [ ] 输入验证是否充分？

---

## 🔄 Pull Request 流程

### 1. 创建 PR

```bash
# 创建功能分支
git checkout -b feat/my-feature

# 提交代码
git commit -m "feat: add my feature"

# 推送到远程
git push origin feat/my-feature
```

### 2. 填写 PR 模板

```markdown
## Description
简要描述变更内容

## Motivation
为什么需要这个变更？

## Changes
- 变更 1
- 变更 2

## Testing
如何测试这些变更？

## Checklist
- [ ] 代码已通过本地测试
- [ ] 已添加必要的单元测试
- [ ] 文档已更新
- [ ] 无 Breaking Changes
```

### 3. Code Review

**作者责任**:
- 回复所有评论
- 解决指出的问题
- 标记已解决的讨论

**审查者责任**:
- 建设性反馈
- 明确指出问题
- 认可优秀实现

### 4. Merge PR

**合并前检查**:
- [ ] 至少 1 个 Approve
- [ ] 所有 CI 检查通过
- [ ] 无未解决的评论
- [ ] 分支是最新的

**合并方式**:
```bash
# Squash and merge（推荐）
# 将多个 commit 压缩为一个

# Rebase and merge
# 保持线性历史

# Create a merge commit
# 保留完整分支历史
```

---

## 📊 分支管理策略

### 分支类型

```
main (保护分支)
  ↑
  ├── develop (开发分支)
  │     ↑
  │     ├── feat/add-new-feature
  │     ├── fix/bug-fix
  │     └── hotfix/critical-fix
  │
  └── release/v1.0.0 (发布分支)
```

### 分支命名

| 类型 | 命名格式 | 示例 |
|------|---------|------|
| 功能分支 | `feat/<description>` | `feat/argo-generator` |
| 修复分支 | `fix/<description>` | `fix/path-separator` |
| 热修复 | `hotfix/<description>` | `hotfix/memory-leak` |
| 发布分支 | `release/<version>` | `release/v1.0.0` |

### 分支生命周期

```bash
# 功能分支：从 develop 创建，完成后合并回 develop
git checkout develop
git checkout -b feat/new-feature
# ... 开发 ...
git checkout develop
git merge feat/new-feature
git branch -d feat/new-feature

# 发布分支：从 develop 分出，稳定后合并到 main
git checkout -b release/v1.0.0
# ... 测试和修复 ...
git checkout main
git merge release/v1.0.0
git tag v1.0.0
```

---

## 🎯 开发工作流

### 日常开发

```bash
# 1. 同步最新代码
git checkout develop
git pull origin develop

# 2. 创建功能分支
git checkout -b feat/my-feature

# 3. 开发和提交
git add .
git commit -m "feat: implement my feature"

# 4. 推送分支
git push origin feat/my-feature

# 5. 创建 Pull Request
# 在 GitHub/GitLab 上创建 PR

# 6. 回应 Code Review
# 根据评论修改代码

# 7. Merge 后清理
git checkout develop
git pull origin develop
git branch -d feat/my-feature
```

### 发布流程

```bash
# 1. 创建发布分支
git checkout -b release/v1.0.0

# 2. 版本测试和修复
# ... 测试 ...

# 3. 更新版本号
# go.mod, version.go 等

# 4. 合并到 main
git checkout main
git merge release/v1.0.0
git tag v1.0.0

# 5. 合并回 develop
git checkout develop
git merge release/v1.0.0

# 6. 删除发布分支
git branch -d release/v1.0.0
```

---

## 📚 相关资源

- [Conventional Commits](https://www.conventionalcommits.org/)
- [GitHub Flow](https://guides.github.com/introduction/flow/)
- [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/)
- [Code Review Guide](https://google.github.io/eng-practices/review/)

---

**最后更新**: 2025-03-14  
**维护者**: K8s App Accelerator Team
