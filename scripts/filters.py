#!/usr/bin/env python3
"""
Ansible 兼容的 Jinja2 Filters 实现

这些 filters 用于保持与原有 Ansible roles 的 100% 兼容性
"""

import logging
from collections import OrderedDict

# 获取 logger
logger = logging.getLogger('render_worker.filters')


def mandatory(value):
    """
    Ansible mandatory filter - 必填校验
    
    Args:
        value: 要检查的值
        
    Returns:
        如果值存在则返回原值，否则抛出异常
        
    Example:
        {{ app_name | mandatory }}
    """
    logger.debug(f"mandatory filter 检查值：{type(value).__name__}")
    if value is None or value == '':
        logger.warning("mandatory filter 检测到空值")
        raise ValueError("mandatory value is required")
    return value


def profile_convert(profile):
    """
    环境名称转换 - 与 Ansible filter_plugins/profile.py 保持完全一致
    
    Args:
        profile: 环境名称
        
    Returns:
        转换后的环境名称
        
    Example:
        {{ profile | profile_convert }}
        int -> Int
        uat -> Uat
        branch -> BRA
        production -> PRD
    """
    if not profile:
        logger.debug("profile_convert 收到空值，返回空字符串")
        return ''
    
    # 与 Ansible filter_plugins/profile.py 完全一致
    mapping = {
        'int': 'Int',
        'uat': 'Uat',
        'branch': 'BRA',
        'production': 'PRD'
    }
    
    result = mapping.get(profile, profile.upper())
    logger.debug(f"profile_convert: {profile} -> {result}")
    return result


def ternary(value, true_val='', false_val=''):
    """
    Ansible ternary filter - 三元运算符
    
    Args:
        value: 条件值
        true_val: 条件为真时的值
        false_val: 条件为假时的值
        
    Returns:
        根据条件返回对应的值
        
    Example:
        {{ enable_hpa | ternary('true', 'false') }}
    """
    if value:
        return true_val
    return false_val


def to_json(value, indent=2, sort_keys=False):
    """
    转换为 JSON 字符串
    
    Args:
        value: 要转换的值
        indent: 缩进空格数
        sort_keys: 是否对字典键排序
        
    Returns:
        JSON 格式字符串
    """
    import json
    logger.debug(f"to_json filter 处理类型：{type(value).__name__}")
    return json.dumps(value, indent=indent, sort_keys=sort_keys)


def combine(dict1, dict2):
    """
    合并两个字典（保持 Ansible 兼容的字典合并顺序）
    
    Ansible 的 combine filter 行为：
    1. 保留 dict1 的键顺序
    2. 添加 dict2 中不存在于 dict1 的键（按 dict2 的顺序）
    
    Args:
        dict1: 第一个字典
        dict2: 第二个字典
        
    Returns:
        合并后的字典
        
    Example:
        {{ default_vars | combine(custom_vars) }}
    """
    logger.debug(f"combine filter: 合并 {len(dict1)} 和 {len(dict2)} 个键值对")
    result = OrderedDict()
    
    # 首先按顺序添加 dict1 的所有键
    for key, value in dict1.items():
        result[key] = value
    
    # 然后按顺序添加 dict2 中不存在于 dict1 的键
    for key, value in dict2.items():
        if key not in result:
            result[key] = value
    
    # Ansible 兼容: 确保常见的键按特定顺序排列
    # jenkins 字典的键顺序: appinstall, rdb_appinstall, rdb, site
    preferred_order = ['appinstall', 'rdb_appinstall', 'rdb', 'site']
    
    # 如果结果中包含这些键，按 preferred_order 排序
    if any(k in result for k in preferred_order):
        # 将 preferred_order 中的键移动到前面
        ordered = OrderedDict()
        for key in preferred_order:
            if key in result:
                ordered[key] = result[key]
        # 添加其他键
        for key, value in result.items():
            if key not in ordered:
                ordered[key] = result[key]
        result = ordered
    
    return dict(result)


def default(value, default_value=''):
    """
    Ansible default filter - 提供默认值
    
    支持两种模式：
    1. 默认模式：default_value 作为默认值
    2. Yes/No 模式：default(yes=true, no=false)
    
    Args:
        value: 要检查的值
        default_value: 默认值
        
    Returns:
        如果值为空则返回默认值，否则返回原值
    """
    # 检查是否是 Jinja2 Undefined 对象
    # Ansible 中，如果变量未定义，会返回 Undefined 对象
    # Jinja2 Undefined 对象有 _undefined_name 属性
    if hasattr(value, '_undefined_name'):
        # 未定义的变量，使用默认值
        logger.debug(f"default filter: 检测到未定义变量 {value._undefined_name}，返回默认值 {default_value}")
        return default_value
    
    # 检查是否是 None 或空字符串
    if value is None or value == '':
        logger.debug(f"default filter: 值为空，返回默认值 {default_value}")
        return default_value
    
    logger.debug(f"default filter: 返回原值 {value}")
    return value


def upper(value):
    """转换为大写"""
    if not value:
        return ''
    return str(value).upper()


def lower(value):
    """转换为小写"""
    if not value:
        return ''
    return str(value).lower()


def first(iterable):
    """
    获取列表/数组的第一个元素
    
    Args:
        iterable: 可迭代对象
        
    Returns:
        第一个元素，如果为空则返回 None
    """
    if iterable and len(iterable) > 0:
        return iterable[0]
    return None


def last(iterable):
    """
    获取列表/数组的最后一个元素
    
    Args:
        iterable: 可迭代对象
        
    Returns:
        最后一个元素，如果为空则返回 None
    """
    if iterable and len(iterable) > 0:
        return iterable[-1]
    return None


def count(iterable):
    """
    获取列表/数组的长度
    
    Args:
        iterable: 可迭代对象
        
    Returns:
        元素个数
    """
    if not iterable:
        return 0
    return len(iterable)


def unique(iterable):
    """
    去重
    
    Args:
        iterable: 可迭代对象
        
    Returns:
        去重后的列表
    """
    if not iterable:
        return []
    seen = set()
    result = []
    for item in iterable:
        if item not in seen:
            seen.add(item)
            result.append(item)
    return result


def difference(list1, list2):
    """
    差集 - list1 中有但 list2 中没有的元素
    
    Args:
        list1: 第一个列表
        list2: 第二个列表
        
    Returns:
        差集列表
    """
    if not list1:
        return []
    if not list2:
        return list1
    
    set2 = set(list2)
    return [item for item in list1 if item not in set2]


def intersect(list1, list2):
    """
    交集
    
    Args:
        list1: 第一个列表
        list2: 第二个列表
        
    Returns:
        交集列表
    """
    if not list1 or not list2:
        return []
    
    set1 = set(list1)
    set2 = set(list2)
    return list(set1 & set2)


def union(list1, list2):
    """
    并集
    
    Args:
        list1: 第一个列表
        list2: 第二个列表
        
    Returns:
        并集列表（去重）
    """
    if not list1:
        return list2 if list2 else []
    if not list2:
        return list1
    
    return list(set(list1) | set(list2))


def flatten(nested_list):
    """
    扁平化嵌套列表
    
    Args:
        nested_list: 嵌套列表 [[1,2], [3,4]] -> [1,2,3,4]
        
    Returns:
        扁平化后的列表
    """
    result = []
    for item in nested_list:
        if isinstance(item, list):
            result.extend(flatten(item))
        else:
            result.append(item)
    return result


def random_filter(value, min_val=0, max_val=None):
    """
    Ansible random filter - 生成随机数
    
    Args:
        value: 最大值（如果 max_val 未指定）或最小值
        min_val: 最小值
        max_val: 最大值（可选）
        
    Returns:
        随机整数
        
    Example:
        {{ 100 | random }}  # 生成 0-99 的随机数
        {{ 20000 | random }}  # 生成 0-19999 的随机数
    """
    if max_val is None:
        # 只提供了一个参数，作为最大值
        max_val = value
        min_val = 0
    
    return random.randint(min_val, max_val - 1)


# ============================================================
# Password Filters - 来自 Ansible filter_plugins/password.py
# ============================================================

import random

# 用于密码生成的字符集
_lower_letters = [chr(i) for i in range(ord('a'), ord('z') + 1)]
_upper_letters = [chr(i) for i in range(ord('A'), ord('Z') + 1)]
_digit_letters = [str(i) for i in range(10)]
_odds_letters = ['!', '*']


def password_generate(origin='', length=12, need_odds=False):
    """
    生成随机密码
    
    Args:
        origin: 原始值（未使用）
        length: 密码长度，最小12
        need_odds: 是否包含特殊字符
        
    Returns:
        生成的密码
        
    Example:
        {{ ''|password_generate(length=16, need_odds=true) }}
    """
    result = ''
    
    # 5个小写字母
    for i in range(5):
        result += random.choice(_lower_letters)
    
    # 1个特殊字符（可选）
    if need_odds:
        result += random.choice(_odds_letters)
    
    # 1个数字
    result += random.choice(_digit_letters)
    
    # 5或6个大写字母
    for i in range(5 if need_odds else 6):
        result += random.choice(_upper_letters)
    
    if length < 12:
        raise ValueError("length parameter of `password_generate` should larger than 12")
    
    # 补充到指定长度
    if length > 12:
        for i in range(length - 12):
            result += random.choice(_lower_letters + _upper_letters + _digit_letters)
    
    return result


def password_gen(origin='', project='miniso', profile='int', length=12):
    """
    基于 project 和 profile 生成密码
    
    Args:
        origin: 原始值（未使用）
        project: 项目名称
        profile: 环境名称
        length: 密码长度
        
    Returns:
        生成的密码
        
    Example:
        {{ ''|password_gen(project='lingfeng', profile='production', length=16) }}
        # Output: Lingfengproduction1AAAAA
    """
    ret = project.capitalize() + profile
    
    if len(ret) >= length:
        if not any(filter(lambda x: x in _digit_letters, ret)):
            return f"{ret[:length - 1]}1"
    
    if not any(filter(lambda x: x in _digit_letters, ret)):
        ret = f"{ret[:-1]}1"
    count = len(ret)
    
    if count < length:
        samples = _upper_letters * 10
        ret += ''.join(samples[:(length - count)])
    
    return ret


def load_filters():
    """
    加载所有 filters
    
    Returns:
        filter 字典，key 是 filter 名称，value 是 filter 函数
    """
    return {
        # 核心 filters
        'mandatory': mandatory,
        'profile_convert': profile_convert,
        'ternary': ternary,
        'to_json': to_json,
        'combine': combine,
        'default': default,
        
        # 字符串处理
        'upper': upper,
        'lower': lower,
        
        # 列表处理
        'first': first,
        'last': last,
        'count': count,
        'unique': unique,
        'difference': difference,
        'intersect': intersect,
        'union': union,
        'flatten': flatten,
        
        # 随机数生成
        'random': random_filter,
        
        # 密码生成 filters（来自 Ansible filter_plugins/password.py）
        'password_generate': password_generate,
        'password_gen': password_gen,
    }


if __name__ == '__main__':
    # 配置基础日志（测试模式）
    logging.basicConfig(
        level=logging.DEBUG,
        format='%(asctime)s [%(levelname)s] %(message)s',
        datefmt='%Y-%m-%d %H:%M:%S'
    )
    
    # 测试 filters
    print("Testing filters...")
    logger.info("开始测试 filters")
    
    # 测试 mandatory
    try:
        assert mandatory('test') == 'test'
        print("✓ mandatory test passed")
    except Exception as e:
        print(f"✗ mandatory test failed: {e}")
    
    # 测试 profile_convert (Ansible 兼容版本)
    assert profile_convert('int') == 'Int'
    assert profile_convert('uat') == 'Uat'
    assert profile_convert('branch') == 'BRA'
    assert profile_convert('production') == 'PRD'
    print("✓ profile_convert test passed (Ansible compatible)")
    
    # 测试 ternary
    assert ternary(True, 'yes', 'no') == 'yes'
    assert ternary(False, 'yes', 'no') == 'no'
    print("✓ ternary test passed")
    
    # 测试 combine
    assert combine({'a': 1}, {'b': 2}) == {'a': 1, 'b': 2}
    print("✓ combine test passed")
    
    # 测试 password_generate
    pwd = password_generate(length=16)
    assert len(pwd) == 16
    print(f"✓ password_generate test passed: {pwd}")
    
    # 测试 password_gen
    pwd2 = password_gen(project='lingfeng', profile='production', length=16)
    assert len(pwd2) >= 16
    print(f"✓ password_gen test passed: {pwd2}")
    
    print("\nAll tests passed!")
