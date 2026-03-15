# K8s App Accelerator Go - 文档中心

本目录包含项目的详细技术文档，按照功能模块和读者角色组织。

---

## 📚 文档导航

### 🎯 快速入口

| 读者角色 | 推荐阅读顺序 | 预计时间 |
|---------|------------|---------|
| **新用户** | 1 → 2 → 3 | 30 分钟 |
| **开发者** | 1 → 4 → 5 → 6 | 60 分钟 |
| **架构师** | 1 → 4 → 9 | 45 分钟 |
| **运维人员** | 1 → 2 → 7 | 40 分钟 |

---

## 📖 文档列表

### 1. [CLI_REFERENCE.md](./CLI_REFERENCE.md) - CLI 命令参考手册 ⭐⭐⭐⭐⭐
**前置条件**: 无  
**目标读者**: 所有用户  
**内容**: 完整的 CLI 命令、Flags、使用示例  
**预计阅读**: 15 分钟

```bash
# 快速查询
go run . --help
```

---

### 2. [DEVELOPMENT_GUIDE.md](./DEVELOPMENT_GUIDE.md) - 开发指南 ⭐⭐⭐⭐⭐
**前置条件**: 无  
**目标读者**: 新用户、开发者  
**内容**: 环境搭建、快速开始、Makefile 使用  
**预计阅读**: 20 分钟

```bash
#  prerequisites
- Go >= 1.21
- Python >= 3.7
- pip3
```

---

### 3. [ARCHITECTURE_DEEP_DIVE.md](./ARCHITECTURE_DEEP_DIVE.md) - 架构深度解析 ⭐⭐⭐⭐
**前置条件**: 
- ✅ 已阅读 [DEVELOPMENT_GUIDE.md](./DEVELOPMENT_GUIDE.md)
- ✅ 了解项目基本结构
- ✅ 熟悉 Go 或 Python 基础

**目标读者**: 开发者、架构师  
**内容**: 核心模块设计、生成器实现、Worker 集成  
**预计阅读**: 40 分钟

```bash
# 需要先理解
- Go 语言基础
- Python 进程通信
- Jinja2 模板
```

---

### 4. [PRECHECK_SPECIFICATION.md](./PRECHECK_SPECIFICATION.md) - Pre-Check 预检规范 ⭐⭐⭐⭐
**前置条件**: 
- ✅ 已阅读 [CLI_REFERENCE.md](./CLI_REFERENCE.md)
- ✅ 了解配置文件结构
- ✅ 运行过 precheck 命令

**目标读者**: 运维人员、开发者  
**内容**: 6 大类检查项、验证规则、错误处理  
**预计阅读**: 25 分钟

```bash
# 实践要求
go run . precheck --base-dir configs
```

---

### 5. [PYTHON_WORKER_IMPLEMENTATION.md](./PYTHON_WORKER_IMPLEMENTATION.md) - Python Worker 实现 ⭐⭐⭐
**前置条件**: 
- ✅ 已阅读 [ARCHITECTURE_DEEP_DIVE.md](./ARCHITECTURE_DEEP_DIVE.md)
- ✅ 理解进程池架构
- ✅ 熟悉 Python 编程

**目标读者**: 核心开发者  
**内容**: Worker 源码、Filters 实现、JSON-RPC 协议  
**预计阅读**: 30 分钟

```bash
# 技术栈
- Python 3.7+
- Jinja2
- JSON-RPC
```

---

### 6. [BEST_PRACTICES.md](./BEST_PRACTICES.md) - 开发最佳实践 ⭐⭐⭐⭐
**前置条件**: 
- ✅ 已完成 [DEVELOPMENT_GUIDE.md](./DEVELOPMENT_GUIDE.md)
- ✅ 有实际开发经验
- ✅ 遇到具体问题

**目标读者**: 开发者  
**内容**: 环境配置、代码规范、调试技巧  
**预计阅读**: 20 分钟

```bash
# 实践经验总结
- 路径处理规范
- 错误处理模式
- 性能优化技巧
```

---

### 7. [ROADMAP.md](./ROADMAP.md) - 项目路线图 ⭐⭐⭐
**前置条件**: 
- ✅ 了解项目背景
- ✅ 关注长期规划

**目标读者**: 架构师、技术负责人  
**内容**: Phase 分解、里程碑、演进方向  
**预计阅读**: 15 分钟

```bash
# 阶段目标
Phase 1: ✅ 核心功能迁移（已完成）
Phase 2: ⏳ 性能优化（进行中）
Phase 3: 📋 模块扩展（规划中）
```

---

### 8. [REFERENCES.md](./REFERENCES.md) - 参考资料 ⭐⭐
**前置条件**: 无  
**目标读者**: 所有用户  
**内容**: 官方文档、技术规范、代码仓库  
**预计阅读**: 5 分钟

```bash
# 外部资源
- Ansible Jinja2 官方文档
- ArgoCD Application 规范
- Cobra CLI 框架
```

---

### 9. [TEAM_COLLABORATION.md](./TEAM_COLLABORATION.md) - 团队协作指南 ⭐⭐⭐
**前置条件**: 
- ✅ 已是项目贡献者
- ✅ 准备提交代码

**目标读者**: 团队成员  
**内容**: Git 规范、Code Review、提交流程  
**预计阅读**: 15 分钟

```bash
# 团队约定
- Git 提交规范
- 代码审查清单
- PR 流程
```

---

## 🔗 与主文档的关系

- **[AGENTS.md](../AGENTS.md)** 是项目的**核心架构文档**
  - 保留最关键的决策和设计
  - 控制在 150 行以内，30 分钟读完
  - **前置条件**: 无，适合快速了解项目

- **docs/ 目录** 是**详细技术文档**
  - 深入实现细节和使用指南
  - 按需阅读，各取所需
  - **前置条件**: 见上方各文档说明

---

## 📊 文档维护

### 更新频率
- **高频率** (每周): BEST_PRACTICES.md, ROADMAP.md
- **中频率** (每月): DEVELOPMENT_GUIDE.md, PRECHECK_SPECIFICATION.md
- **低频率** (每季度): ARCHITECTURE_DEEP_DIVE.md, CLI_REFERENCE.md

### 维护责任
- **文档作者**: 负责初始版本
- **模块负责人**: 负责持续更新
- **Tech Lead**: 负责最终审阅

### 版本控制
所有文档在文件末尾标注：
- 最后更新日期
- 版本号
- 主要贡献者

---

## 🎯 文档使用建议

### 第一次阅读
1. 先读 [AGENTS.md](../AGENTS.md) 了解全貌
2. 根据角色选择 2-3 个重点文档
3. 遇到问题时查阅相关章节

### 日常开发
1. 以 [BEST_PRACTICES.md](./BEST_PRACTICES.md) 为指导
2. 参考 [CLI_REFERENCE.md](./CLI_REFERENCE.md) 编写命令
3. 遵循 [TEAM_COLLABORATION.md](./TEAM_COLLABORATION.md) 提交代码

### 架构设计
1. 深入理解 [ARCHITECTURE_DEEP_DIVE.md](./ARCHITECTURE_DEEP_DIVE.md)
2. 参考 [ROADMAP.md](./ROADMAP.md) 规划演进
3. 查阅 [REFERENCES.md](./REFERENCES.md) 获取灵感

---

**文档中心最后更新**: 2025-03-14  
**维护者**: K8s App Accelerator Team
