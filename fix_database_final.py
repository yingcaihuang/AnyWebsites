#!/usr/bin/env python3
"""
最终修复database-init.sql中的所有问题
1. 修复contents表中的重复ID
2. 确保content_analytics表中的content_id引用正确
"""

import re

def fix_database_final():
    # 读取文件
    with open('database-init.sql', 'r', encoding='utf-8') as f:
        content = f.read()
    
    # 1. 修复contents表中的重复ID和不连续ID
    contents_section = """-- 插入测试内容
INSERT INTO contents (id, user_id, title, content, content_type, access_count, created_at)
VALUES
    ('10000001-1111-1111-1111-111111111111', '6d53e189-76e0-4c4a-b94d-4e942e25bf60', '欢迎使用 AnyWebsites', '这是一个示例文本内容，展示了 AnyWebsites 的基本功能。您可以上传文本、图片、文档等各种类型的文件。', 'text/plain', 156, CURRENT_TIMESTAMP - INTERVAL '7 days'),
    ('10000002-2222-2222-2222-222222222222', '56ecc01e-9ffc-4515-8932-19912ce0805d', 'API 使用指南', '# API 使用指南\\n\\n## 认证\\n使用 Bearer Token 进行认证：\\n```\\nAuthorization: Bearer YOUR_API_KEY\\n```\\n\\n## 上传文件\\n```bash\\ncurl -X POST https://api.anywebsites.com/api/content/upload \\\\\\n  -H "Authorization: Bearer YOUR_API_KEY" \\\\\\n  -F "file=@example.txt"\\n```', 'text/markdown', 89, CURRENT_TIMESTAMP - INTERVAL '5 days'),
    ('10000003-3333-3333-3333-333333333333', 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', '测试数据', '这是一些测试数据，用于演示系统功能。', 'text/plain', 23, CURRENT_TIMESTAMP - INTERVAL '3 days'),
    ('10000004-4444-4444-4444-444444444444', 'bbbbbbbb-cccc-dddd-eeee-ffffffffffff', 'JSON 配置示例', '{"name": "example", "version": "1.0.0", "description": "示例配置文件"}', 'application/json', 45, CURRENT_TIMESTAMP - INTERVAL '2 days'),
    ('10000005-5555-5555-5555-555555555555', '6d53e189-76e0-4c4a-b94d-4e942e25bf60', '系统公告', '系统将于本周末进行维护升级，预计停机时间为2小时。感谢您的理解与支持！', 'text/plain', 234, CURRENT_TIMESTAMP - INTERVAL '1 day');"""
    
    # 2. 修复content_analytics表中的content_id引用
    analytics_section = """-- 插入内容分析数据（模拟访问记录）
INSERT INTO content_analytics (id, content_id, ip_address, user_agent, referer, country, region, city, access_time)
VALUES
    -- 中国访问记录
    ('10000001-1111-1111-1111-111111111111', '10000001-1111-1111-1111-111111111111', '192.168.1.100', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', 'https://google.com', 'China', 'Beijing', 'Beijing', CURRENT_TIMESTAMP - INTERVAL '6 days'),
    ('10000002-2222-2222-2222-222222222222', '10000001-1111-1111-1111-111111111111', '192.168.1.101', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36', 'https://baidu.com', 'China', 'Shanghai', 'Shanghai', CURRENT_TIMESTAMP - INTERVAL '6 days'),
    ('10000003-3333-3333-3333-333333333333', '10000002-2222-2222-2222-222222222222', '192.168.1.102', 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X)', 'https://github.com', 'China', 'Guangdong', 'Shenzhen', CURRENT_TIMESTAMP - INTERVAL '5 days'),

    -- 美国访问记录
    ('10000004-4444-4444-4444-444444444444', '10000001-1111-1111-1111-111111111111', '203.0.113.1', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', 'https://google.com', 'United States', 'California', 'San Francisco', CURRENT_TIMESTAMP - INTERVAL '4 days'),
    ('10000005-5555-5555-5555-555555555555', '10000002-2222-2222-2222-222222222222', '203.0.113.2', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36', 'https://stackoverflow.com', 'United States', 'New York', 'New York', CURRENT_TIMESTAMP - INTERVAL '4 days'),

    -- 日本访问记录
    ('10000006-6666-6666-6666-666666666666', '10000003-3333-3333-3333-333333333333', '198.51.100.1', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', 'https://yahoo.co.jp', 'Japan', 'Tokyo', 'Tokyo', CURRENT_TIMESTAMP - INTERVAL '3 days'),

    -- 德国访问记录
    ('10000007-7777-7777-7777-777777777777', '10000004-4444-4444-4444-444444444444', '198.51.100.2', 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36', 'https://google.de', 'Germany', 'Bavaria', 'Munich', CURRENT_TIMESTAMP - INTERVAL '2 days'),

    -- 英国访问记录
    ('10000008-8888-8888-8888-888888888888', '10000005-5555-5555-5555-555555555555', '198.51.100.3', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', 'https://bbc.co.uk', 'United Kingdom', 'England', 'London', CURRENT_TIMESTAMP - INTERVAL '1 day'),

    -- 更多中国访问记录
    ('10000009-9999-9999-9999-999999999999', '10000001-1111-1111-1111-111111111111', '192.168.1.103', 'Mozilla/5.0 (Android 11; Mobile) AppleWebKit/537.36', 'https://weibo.com', 'China', 'Zhejiang', 'Hangzhou', CURRENT_TIMESTAMP - INTERVAL '1 day'),
    ('1000000a-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '10000002-2222-2222-2222-222222222222', '192.168.1.104', 'Mozilla/5.0 (iPad; CPU OS 14_7_1 like Mac OS X)', 'https://zhihu.com', 'China', 'Jiangsu', 'Nanjing', CURRENT_TIMESTAMP - INTERVAL '12 hours');"""
    
    # 替换contents部分
    contents_pattern = r"-- 插入测试内容.*?CURRENT_TIMESTAMP - INTERVAL '1 day'\);"
    content = re.sub(contents_pattern, contents_section, content, flags=re.DOTALL)
    
    # 替换content_analytics部分
    analytics_pattern = r"-- 插入内容分析数据.*?CURRENT_TIMESTAMP - INTERVAL '12 hours'\);"
    content = re.sub(analytics_pattern, analytics_section, content, flags=re.DOTALL)
    
    # 写回文件
    with open('database-init.sql', 'w', encoding='utf-8') as f:
        f.write(content)
    
    print("数据库初始化文件最终修复完成！")
    print("- 修复了contents表中的重复ID")
    print("- 统一了所有content_id引用")
    print("- 确保外键约束正确")

if __name__ == '__main__':
    fix_database_final()
