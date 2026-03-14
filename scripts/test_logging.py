#!/usr/bin/env python3
"""
日志功能测试脚本

用于验证 render_worker.py 的日志系统是否正常工作
"""

import sys
import os
import json
import tempfile
import subprocess
from datetime import datetime


def test_logging_setup():
    """测试日志配置"""
    print("=" * 60)
    print("测试 1: 日志配置")
    print("=" * 60)
    
    # 设置不同的日志级别
    test_levels = ['DEBUG', 'INFO', 'WARNING', 'ERROR']
    
    for level in test_levels:
        print(f"\n测试 LOG_LEVEL={level}")
        os.environ['LOG_LEVEL'] = level
        
        # 运行测试模式
        result = subprocess.run(
            ['python3', 'render_worker.py'],
            input='test',
            capture_output=True,
            text=True
        )
        
        # 检查 stderr 中是否有日志输出
        if result.stderr:
            print(f"✓ 日志输出正常 ({len(result.stderr)} bytes)")
            print(f"  示例：{result.stderr.split(chr(10))[0]}")
        else:
            print(f"✗ 无日志输出")
    
    print("\n")


def test_template_rendering():
    """测试模板渲染日志"""
    print("=" * 60)
    print("测试 2: 模板渲染日志")
    print("=" * 60)
    
    # 创建临时模板
    with tempfile.NamedTemporaryFile(mode='w', suffix='.j2', delete=False) as f:
        f.write("""
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ app_name }}
  namespace: {{ namespace | default("default") }}
data:
  key: {{ value }}
""")
        template_path = f.name
    
    try:
        # 测试模式渲染
        context = {
            'app_name': 'test-app',
            'namespace': 'production',
            'value': 'test-value'
        }
        
        print(f"\n渲染模板：{template_path}")
        print(f"上下文：{json.dumps(context)}")
        
        os.environ['LOG_LEVEL'] = 'DEBUG'
        
        result = subprocess.run(
            ['python3', 'render_worker.py', template_path, json.dumps(context)],
            capture_output=True,
            text=True
        )
        
        print(f"\nstdout:\n{result.stdout}")
        print(f"\nstderr (日志):\n{result.stderr}")
        
        if result.returncode == 0:
            print("✓ 模板渲染成功")
        else:
            print(f"✗ 模板渲染失败 (returncode={result.returncode})")
    
    finally:
        # 清理临时文件
        os.unlink(template_path)
    
    print("\n")


def test_error_handling():
    """测试错误处理日志"""
    print("=" * 60)
    print("测试 3: 错误处理日志")
    print("=" * 60)
    
    # 测试不存在的模板
    print("\n测试：渲染不存在的模板")
    
    os.environ['LOG_LEVEL'] = 'DEBUG'
    
    result = subprocess.run(
        ['python3', 'render_worker.py', '/nonexistent/template.yaml.j2', '{}'],
        capture_output=True,
        text=True
    )
    
    print(f"\nstderr (错误日志):\n{result.stderr}")
    
    if 'ERROR' in result.stderr and '模板文件不存在' in result.stderr:
        print("✓ 错误日志记录正确")
    else:
        print("✗ 错误日志记录不完整")
    
    print("\n")


def test_filters_logging():
    """测试 filters 日志"""
    print("=" * 60)
    print("测试 4: Filters 日志")
    print("=" * 60)
    
    # 创建使用 filters 的模板
    with tempfile.NamedTemporaryFile(mode='w', suffix='.j2', delete=False) as f:
        f.write("""
mandatory: {{ mandatory_value | mandatory }}
profile: {{ profile | profile_convert }}
ternary: {{ enable | ternary('enabled', 'disabled') }}
combined: {{ (dict1 | combine(dict2)).key1 }}
""")
        template_path = f.name
    
    try:
        context = {
            'mandatory_value': 'required',
            'profile': 'int',
            'enable': True,
            'dict1': {'key1': 'value1', 'key2': 'value2'},
            'dict2': {'key2': 'updated', 'key3': 'value3'}
        }
        
        print(f"\n使用 filters 的模板：{template_path}")
        
        os.environ['LOG_LEVEL'] = 'DEBUG'
        
        result = subprocess.run(
            ['python3', 'render_worker.py', template_path, json.dumps(context)],
            capture_output=True,
            text=True
        )
        
        print(f"\nstderr (filters 日志):\n{result.stderr}")
        
        # 检查是否有 filter 相关的日志
        if 'profile_convert' in result.stderr or 'combine' in result.stderr:
            print("✓ Filters 日志记录正常")
        else:
            print("⚠ Filters 日志可能未完全启用")
        
        print(f"\n渲染结果:\n{result.stdout}")
    
    finally:
        os.unlink(template_path)
    
    print("\n")


def test_mandatory_error():
    """测试 mandatory filter 错误日志"""
    print("=" * 60)
    print("测试 5: Mandatory Filter 错误日志")
    print("=" * 60)
    
    with tempfile.NamedTemporaryFile(mode='w', suffix='.j2', delete=False) as f:
        f.write("{{ empty_value | mandatory }}")
        template_path = f.name
    
    try:
        context = {'empty_value': ''}
        
        print("\n测试：mandatory filter 检测空值")
        
        os.environ['LOG_LEVEL'] = 'DEBUG'
        
        result = subprocess.run(
            ['python3', 'render_worker.py', template_path, json.dumps(context)],
            capture_output=True,
            text=True
        )
        
        print(f"\nstderr (错误日志):\n{result.stderr}")
        
        if 'WARNING' in result.stderr and 'mandatory' in result.stderr:
            print("✓ Mandatory filter 警告记录正确")
        else:
            print("⚠ Mandatory filter 警告可能未记录")
    
    finally:
        os.unlink(template_path)
    
    print("\n")


def main():
    """主函数"""
    print("\n")
    print("╔" + "=" * 58 + "╗")
    print("║" + " " * 15 + "Render Worker 日志功能测试" + " " * 15 + "║")
    print("╚" + "=" * 58 + "╝")
    print()
    
    # 切换到 scripts 目录
    script_dir = os.path.dirname(os.path.abspath(__file__))
    os.chdir(script_dir)
    
    print(f"工作目录：{os.getcwd()}")
    print(f"Python 版本：{sys.version}")
    print()
    
    start_time = datetime.now()
    
    try:
        # 运行所有测试
        test_logging_setup()
        test_template_rendering()
        test_error_handling()
        test_filters_logging()
        test_mandatory_error()
        
        end_time = datetime.now()
        duration = (end_time - start_time).total_seconds()
        
        print("=" * 60)
        print(f"所有测试完成，耗时：{duration:.2f}秒")
        print("=" * 60)
        
    except KeyboardInterrupt:
        print("\n\n测试被用户中断")
    except Exception as e:
        print(f"\n✗ 测试过程中发生异常：{e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == '__main__':
    main()
