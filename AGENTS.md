# K8s App Accelerator Go - 核心架构文档

## 📋 项目概述

### 项目背景
将现有的 Ansible-based K8s 配置生成器迁移到 Golang实现，保持 100% Jinja2 模板兼容性，同时获得 Go语言的性能、类型安全和工程化优势。

**当前阶段**: 已完成 ArgoCD、Jenkins、CMDB 模块的迁移  
**长期目标**: 逐步迁移整个 `k8s_app_acelerator` 下的所有功能模块

### 核心目标
- ✅ **100% 兼容现有 Jinja2 模板**（无需修改 Ansible roles）
- ✅ **性能提升 5 倍以上**（从 8 分钟缩短到 1.5 分钟）
- ✅ **跨平台支持**（Windows/macOS/Linux原生运行）
- ✅ **智能预检功能**（减少 80% 配置错误）
- ✅ **开发友好**（支持 `go run` 直接运行）

---

## 🏗️ 技术架构

### 方案选择：Go 主程序 + Python 子进程

```
┌─────────────────────────────────────┐
│         Golang (主程序)             │
│  - 配置加载 (YAML 解析)              │
│  - 流程编排                          │
│  - Pre-Check 预检                   │
│  - CLI 交互                         │
│  - 文件写入                          │
└──────────────┬──────────────────────┘
               │ JSON-RPC over stdin/stdout
               │ 进程池管理 (5 workers)
               ↓
┌─────────────────────────────────────┐
│      Python Worker Pool             │
│  ┌──────────────────────────────┐   │
│  │   Jinja2 Template Engine     │   │
│  │   - 加载模板                  │   │
│  │   - 渲染数据                  │   │
│  │   - 返回结果                  │   │
│  └──────────────────────────────┘   │
└─────────────────────────────────────┘
```

### 为什么选择这个方案？

| 维度 | 评估 | 说明 |
|------|------|------|
| **兼容性** | ⭐⭐⭐⭐⭐ | 100% 兼容现有 Jinja2 模板和 Ansible filters |
| **性能** | ⭐⭐⭐⭐ | 进程池优化后达到 5 倍提升 |
| **开发效率** | ⭐⭐⭐⭐⭐ | Go 负责逻辑，Python 专注渲染 |
| **部署** | ⭐⭐⭐⭐ | 单二进制 + Python 运行时 |
| **维护成本** | ⭐⭐⭐⭐ | 清晰的模块边界 |
| **跨平台** | ⭐⭐⭐⭐⭐ | Windows/macOS/Linux原生支持，无需 WSL 或虚拟机 |

---

## 📚 文档导航

本文档仅包含项目的**核心架构设计和关键决策**。详细内容请查阅 [docs/](./docs/) 目录：

### 🎯 快速入口

| 读者角色 | 推荐阅读 | 前置条件 |
|---------|---------|---------|
| **新用户** | [开发指南](./docs/DEVELOPMENT_GUIDE.md) → [CLI 参考](./docs/CLI_REFERENCE.md) | 无 |
| **开发者** | [架构解析](./docs/ARCHITECTURE_DEEP_DIVE.md) → [最佳实践](./docs/BEST_PRACTICES.md) | 已阅读开发指南 |
| **架构师** | [架构解析](./docs/ARCHITECTURE_DEEP_DIVE.md) → [路线图](./docs/ROADMAP.md) | 了解项目背景 |
| **运维人员** | [CLI 参考](./docs/CLI_REFERENCE.md) → [预检规范](./docs/PRECHECK_SPECIFICATION.md) | 无 |

### 📖 完整文档列表

1. **[CLI_REFERENCE.md](./docs/CLI_REFERENCE.md)** - CLI 命令参考手册（130 行）  
   **前置条件**: 无

2. **[DEVELOPMENT_GUIDE.md](./docs/DEVELOPMENT_GUIDE.md)** - 开发指南（330 行）  
   **前置条件**: 无

3. **[ARCHITECTURE_DEEP_DIVE.md](./docs/ARCHITECTURE_DEEP_DIVE.md)** - 架构深度解析（560 行）  
   **前置条件**: 已阅读开发指南，熟悉 Go/Python 基础

4. **[PRECHECK_SPECIFICATION.md](./docs/PRECHECK_SPECIFICATION.md)** - Pre-Check 预检规范（630 行）  
   **前置条件**: 已阅读 CLI 参考，了解配置文件结构

5. **[PYTHON_WORKER_IMPLEMENTATION.md](./docs/PYTHON_WORKER_IMPLEMENTATION.md)** - Python Worker 实现（380 行）  
   **前置条件**: 已阅读架构解析，理解进程池架构

6. **[BEST_PRACTICES.md](./docs/BEST_PRACTICES.md)** - 开发最佳实践（370 行）  
   **前置条件**: 有实际开发经验

7. **[ROADMAP.md](./docs/ROADMAP.md)** - 项目路线图（290 行）  
   **前置条件**: 了解项目背景

8. **[REFERENCES.md](./docs/REFERENCES.md)** - 参考资料（200 行）  
   **前置条件**: 无

9. **[TEAM_COLLABORATION.md](./docs/TEAM_COLLABORATION.md)** - 团队协作指南（310 行）  
   **前置条件**: 已是项目贡献者

**总计**: 约 3200 行详细技术文档

---

## 📊 性能基准

| 场景 | Ansible | Go + Python | 提升 |
|------|---------|-------------|------|
| **启动时间** | ~500ms | ~50ms | **10x** |
| **单个应用生成** | ~2-3 秒 | ~0.3-0.5 秒 | **6x** |
| **100 个应用全量生成** | 3 分 30 秒 | 45 秒 | **4.7x** |
| **内存占用** | 300-500MB | 50-100MB | **70%** |

**关键指标**:
- 进程池大小：5 个 workers（经验值）
- 并发控制：限制最大并发数为 10
- 超时设置：单次渲染超时 30 秒
- 重试机制：失败自动重试 2 次

---

## 🔮 未来演进路线

### Phase 1: 核心功能迁移（✅ 已完成）
- Go + Python 子进程架构
- ArgoCD/Jenkins/CMDB 生成器
- Pre-Check 预检功能

### Phase 2: 性能优化（⏳ 进行中）
- 动态 Worker 池优化
- 缓存机制
- 内存泄漏检测

### Phase 3: 模块扩展（📋 规划中）
- 应用管理工具
- Stack 管理工具
- 监控运维工具

### Phase 4: 生态建设（🔮 愿景）
- Web UI
- 插件系统
- CI/CD 集成

详见：[ROADMAP.md](./docs/ROADMAP.md)

---

## 🎉 成功标准

项目成功的标志：

1. ✅ **零模板修改** - 现有 Jinja2 模板无需任何改动
2. ✅ **性能达标** - 全量生成时间 < 2 分钟
3. ✅ **用户满意** - 运维人员 30 分钟内上手
4. ✅ **稳定可靠** - 生产环境零故障
5. ✅ **易于维护** - 新人 1 周内可贡献代码

---

## 🔗 快速链接

- **[文档中心](./docs/README.md)** - 完整技术文档索引
- **[GitHub 仓库](https://github.com/buhaiqing/k8s-app-accelerator-go)** - 源代码
- **[Ansible 版本](https://github.com/buhaiqing/k8s_app_acelerator)** - 原始实现

---

**最后更新**: 2025-03-14  
**维护者**: K8s App Accelerator Team  
**文档版本**: v2.0 (重构版)
