#!/usr/bin/env python3
"""
修复database-init.sql中的外键引用问题
确保content_analytics表中的content_id与contents表中的id匹配
"""

import re

def fix_foreign_keys():
    # 读取文件
    with open('database-init.sql', 'r', encoding='utf-8') as f:
        content = f.read()
    
    # 查找contents表的插入语句，提取实际的content IDs
    contents_pattern = r"INSERT INTO contents.*?VALUES\s*\((.*?)\);"
    contents_match = re.search(contents_pattern, content, re.DOTALL | re.IGNORECASE)
    
    if contents_match:
        contents_values = contents_match.group(1)
        # 提取所有content IDs (第一个字段)
        content_ids = []
        lines = contents_values.split('\n')
        for line in lines:
            line = line.strip()
            if line.startswith("('") and "," in line:
                # 提取第一个UUID
                uuid_match = re.match(r"\s*\('([^']+)'", line)
                if uuid_match:
                    content_ids.append(uuid_match.group(1))
        
        print(f"找到的content IDs: {content_ids}")
        
        # 如果找到了content IDs，更新content_analytics表中的引用
        if content_ids:
            # 定义content_analytics中应该使用的content_id映射
            analytics_replacements = {
                '20000001-1111-1111-1111-111111111111': content_ids[0] if len(content_ids) > 0 else '20000001-1111-1111-1111-111111111111',
                '20000002-2222-2222-2222-222222222222': content_ids[1] if len(content_ids) > 1 else '20000002-2222-2222-2222-222222222222',
                '20000003-3333-3333-3333-333333333333': content_ids[2] if len(content_ids) > 2 else '20000003-3333-3333-3333-333333333333',
                '20000004-4444-4444-4444-444444444444': content_ids[3] if len(content_ids) > 3 else '20000004-4444-4444-4444-444444444444',
                '20000005-5555-5555-5555-555555555555': content_ids[4] if len(content_ids) > 4 else '20000005-5555-5555-5555-555555555555',
            }
            
            # 执行替换
            for old_id, new_id in analytics_replacements.items():
                if old_id != new_id:
                    content = content.replace(old_id, new_id)
                    print(f"替换: {old_id} -> {new_id}")
    
    # 另一种方法：直接统一所有的content IDs
    # 确保contents表和content_analytics表使用相同的UUID格式
    
    # 统一的content IDs
    unified_content_ids = [
        '10000001-1111-1111-1111-111111111111',
        '10000002-2222-2222-2222-222222222222', 
        '10000003-3333-3333-3333-333333333333',
        '10000004-4444-4444-4444-444444444444',
        '10000005-5555-5555-5555-555555555555'
    ]
    
    # 在contents表中使用统一的IDs
    old_content_patterns = [
        r"'[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'(?=,\s*'[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}',\s*'欢迎使用 AnyWebsites')",
        r"'[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'(?=,\s*'[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}',\s*'API 使用指南')",
        r"'[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'(?=,\s*'[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}',\s*'测试数据')",
        r"'[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'(?=,\s*'[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}',\s*'JSON 配置示例')",
        r"'[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'(?=,\s*'[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}',\s*'系统公告')"
    ]
    
    # 简单的方法：直接替换所有已知的错误引用
    error_mappings = {
        # content_analytics中错误的content_id引用
        "'20000001-1111-1111-1111-111111111111'": "'10000001-1111-1111-1111-111111111111'",
        "'20000002-2222-2222-2222-222222222222'": "'10000002-2222-2222-2222-222222222222'",
        "'20000003-3333-3333-3333-333333333333'": "'10000003-3333-3333-3333-333333333333'",
        "'20000004-4444-4444-4444-444444444444'": "'10000004-4444-4444-4444-444444444444'",
        "'20000005-5555-5555-5555-555555555555'": "'10000005-5555-5555-5555-555555555555'",
    }
    
    for old_ref, new_ref in error_mappings.items():
        if old_ref in content:
            content = content.replace(old_ref, new_ref)
            print(f"修复外键引用: {old_ref} -> {new_ref}")
    
    # 写回文件
    with open('database-init.sql', 'w', encoding='utf-8') as f:
        f.write(content)
    
    print("外键引用修复完成！")

if __name__ == '__main__':
    fix_foreign_keys()
