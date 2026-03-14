# Python Worker 实现详解

**前置条件**: 
- ✅ 已阅读 [ARCHITECTURE_DEEP_DIVE.md](./ARCHITECTURE_DEEP_DIVE.md)
- ✅ 理解进程池架构
- ✅ 熟悉 Python 编程

**目标读者**: 核心开发者  
**最后更新**: 2025-03-14

---

## 🎯 架构设计

### 为什么需要 Python Worker？

Go 语言无法直接运行 Jinja2 模板，需要通过子进程调用 Python。

```
┌─────────────┐         ┌──────────────┐
│   Go Main   │  JSON   │ Python Worker│
│  Process    │ ←────→  │  (Jinja2)    │
│             │  stdin  │              │
└─────────────┘         └──────────────┘
```

---

## 💻 完整代码实现

### render_worker.py

```python
#!/usr/bin/env python3
"""
Jinja2 渲染 Worker - 支持 JSON-RPC 通信
"""

import sys
import json
from jinja2 import Environment, FileSystemLoader

def load_filters():
    """加载 Ansible 兼容的 filters"""
    
    def ternary(value, true_val='', false_val=''):
        """Ansible ternary filter"""
        return true_val if value else false_val
    
    def profile_convert(profile):
        """int -> INT, production -> PRODUCTION"""
        return profile.upper()
    
    def mandatory(value):
        """必填校验"""
        if not value:
            raise ValueError("mandatory value is required")
        return value
    
    def to_json(value):
        """转换为 JSON 字符串"""
        return json.dumps(value)
    
    def from_json(value):
        """从 JSON 字符串解析"""
        return json.loads(value)
    
    return {
        'ternary': ternary,
        'upper': str.upper,
        'lower': str.lower,
        'profile_convert': profile_convert,
        'mandatory': mandatory,
        'to_json': to_json,
        'from_json': from_json,
    }

def main():
    # 初始化 Jinja2 环境
    env = Environment(loader=FileSystemLoader('/'))
    env.filters.update(load_filters())
    
    # Worker 模式：持续读取 stdin
    if len(sys.argv) > 1 and sys.argv[1] == '--worker-mode':
        while True:
            try:
                line = sys.stdin.readline()
                if not line:
                    break
                
                req = json.loads(line.strip())
                template_path = req['template_path']
                context = req['context']
                
                template = env.get_template(template_path)
                result = template.render(**context)
                
                # 返回 JSON 响应
                resp = {'content': result}
                print(json.dumps(resp), flush=True)
                
            except Exception as e:
                resp = {'error': str(e)}
                print(json.dumps(resp), flush=True)

if __name__ == '__main__':
    main()
```

---

## 🔧 Filters 实现

### filters.py

```python
"""
Ansible-compatible Jinja2 filters implementation
"""

def ternary(value, true_val='', false_val=''):
    """
    Ansible ternary filter
    
    Usage:
      {{ condition | ternary('yes', 'no') }}
    """
    return true_val if value else false_val


def profile_convert(profile):
    """
    Convert profile name to uppercase
    
    Usage:
      {{ 'int' | profile_convert }} => 'INT'
    """
    return profile.upper()


def mandatory(value):
    """
    Mandatory value check
    
    Usage:
      {{ value | mandatory }}
    """
    if not value:
        raise ValueError("mandatory value is required")
    return value


def to_json(value):
    """
    Convert to JSON string
    
    Usage:
      {{ dict | to_json }}
    """
    return json.dumps(value, indent=2)


def from_json(value):
    """
    Parse JSON string
    
    Usage:
      {{ json_string | from_json }}
    """
    return json.loads(value)


def boolFromString(value):
    """
    Convert string to boolean
    
    Usage:
      {{ 'true' | boolFromString }} => True
    """
    if isinstance(value, bool):
        return value
    return value.lower() in ('true', 'yes', '1', 'on')
```

---

## 📡 JSON-RPC 协议

### 请求格式

```json
{
  "template_path": "templates/argo-app/app.yaml.j2",
  "context": {
    "project": "dly",
    "profile": "production",
    "stack": "zt4d",
    "app": "gateway-service"
  }
}
```

### 响应格式（成功）

```json
{
  "content": "apiVersion: argoproj.io/v1alpha1\nkind: Application\n..."
}
```

### 响应格式（失败）

```json
{
  "error": "Template not found: templates/argo-app/app.yaml.j2"
}
```

---

## 🔄 Worker 生命周期

### 1. 启动

```go
cmd := exec.Command("python3", "scripts/render_worker.py", "--worker-mode")
stdin, _ := cmd.StdinPipe()
stdout, _ := cmd.StdoutPipe()
cmd.Start()
```

### 2. 发送请求

```go
req := RenderRequest{
    TemplatePath: "app.yaml.j2",
    Context: ctx,
}
encoder := json.NewEncoder(stdin)
encoder.Encode(req)
```

### 3. 接收响应

```go
decoder := json.NewDecoder(stdout)
var resp RenderResponse
decoder.Decode(&resp)
```

### 4. 关闭 Worker

```go
cmd.Process.Kill()
```

---

## ⚠️ 错误处理

### 常见错误及解决方案

#### 错误 1: Template not found

**原因**: 模板文件路径不正确  
**解决**: 使用绝对路径或检查工作目录

```python
try:
    template = env.get_template(template_path)
except jinja2.exceptions.TemplateNotFound:
    resp = {'error': f'Template not found: {template_path}'}
```

#### 错误 2: Undefined variable

**原因**: 渲染上下文中缺少变量  
**解决**: 检查 context 是否包含所有必需变量

```python
try:
    result = template.render(**context)
except jinja2.exceptions.UndefinedError as e:
    resp = {'error': f'Undefined variable: {str(e)}'}
```

#### 错误 3: JSON decode error

**原因**: stdin 数据不是有效的 JSON  
**解决**: 验证请求格式

```python
try:
    req = json.loads(line.strip())
except json.JSONDecodeError:
    resp = {'error': 'Invalid JSON request'}
```

---

## 📊 性能优化

### 1. 进程复用

避免频繁启动/停止 Worker：

```python
# Worker 模式：持续运行
while True:
    line = sys.stdin.readline()
    if not line:
        break
    # 处理请求...
```

### 2. 并发控制

限制最大并发 Worker 数量：

```go
// 推荐：5 个 Workers
pool := NewWorkerPool(5, scriptPath)
```

### 3. 超时设置

防止单个渲染卡死：

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

---

## 🧪 测试方法

### 手动测试 Worker

```bash
# 启动 Worker
python3 scripts/render_worker.py --worker-mode

# 发送测试请求（另开终端）
echo '{"template_path": "test.j2", "context": {"name": "test"}}' | nc localhost 12345
```

### 单元测试

```python
import unittest

class TestFilters(unittest.TestCase):
    
    def test_ternary(self):
        self.assertEqual(ternary(True, 'yes', 'no'), 'yes')
        self.assertEqual(ternary(False, 'yes', 'no'), 'no')
    
    def test_profile_convert(self):
        self.assertEqual(profile_convert('int'), 'INT')
        self.assertEqual(profile_convert('production'), 'PRODUCTION')
    
    def test_mandatory(self):
        self.assertEqual(mandatory('value'), 'value')
        with self.assertRaises(ValueError):
            mandatory('')

if __name__ == '__main__':
    unittest.main()
```

---

## 📚 相关文档

- [ARCHITECTURE_DEEP_DIVE.md](./ARCHITECTURE_DEEP_DIVE.md) - 整体架构设计
- [BEST_PRACTICES.md](./BEST_PRACTICES.md) - 开发最佳实践
- [DEVELOPMENT_GUIDE.md](./DEVELOPMENT_GUIDE.md) - 环境搭建指南

---

**最后更新**: 2025-03-14  
**维护者**: K8s App Accelerator Team
