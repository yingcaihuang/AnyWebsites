-- 启用 UUID 扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 删除所有表（如果存在）
DROP TABLE IF EXISTS content_analytics CASCADE;
DROP TABLE IF EXISTS user_subscriptions CASCADE;
DROP TABLE IF EXISTS usage_statistics CASCADE;
DROP TABLE IF EXISTS plan_upgrade_history CASCADE;
DROP TABLE IF EXISTS system_setting_history CASCADE;
DROP TABLE IF EXISTS system_settings CASCADE;
DROP TABLE IF EXISTS system_setting_categories CASCADE;
DROP TABLE IF EXISTS contents CASCADE;
DROP TABLE IF EXISTS plan_configs CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- 创建用户表
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    api_key VARCHAR(255) UNIQUE,
    is_admin BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建计划配置表
CREATE TABLE plan_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    max_uploads_per_day INTEGER DEFAULT 10,
    max_file_size_mb INTEGER DEFAULT 10,
    max_storage_mb INTEGER DEFAULT 100,
    retention_days INTEGER DEFAULT 30,
    price_monthly DECIMAL(10,2) DEFAULT 0.00,
    price_yearly DECIMAL(10,2) DEFAULT 0.00,
    features JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建内容表
CREATE TABLE contents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255),
    content TEXT NOT NULL,
    content_type VARCHAR(50) DEFAULT 'text/plain',
    file_path VARCHAR(500),
    file_size BIGINT DEFAULT 0,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    deleted_at TIMESTAMP,
    access_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建内容分析表
CREATE TABLE content_analytics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content_id UUID NOT NULL REFERENCES contents(id) ON DELETE CASCADE,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    referer TEXT,
    country VARCHAR(100),
    region VARCHAR(100),
    city VARCHAR(100),
    access_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建用户订阅表
CREATE TABLE user_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plan_configs(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'active',
    starts_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    auto_renew BOOLEAN DEFAULT FALSE,
    payment_method VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建使用统计表
CREATE TABLE usage_statistics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    month_year VARCHAR(7) NOT NULL, -- 格式: YYYY-MM
    uploads_count INTEGER DEFAULT 0,
    storage_used_mb INTEGER DEFAULT 0,
    api_calls_count INTEGER DEFAULT 0,
    bandwidth_used_mb INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, month_year)
);

-- 创建计划升级历史表
CREATE TABLE plan_upgrade_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_plan_id UUID REFERENCES plan_configs(id),
    to_plan_id UUID NOT NULL REFERENCES plan_configs(id),
    upgrade_type VARCHAR(20) NOT NULL, -- 'upgrade', 'downgrade', 'change'
    reason TEXT,
    effective_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建系统设置分类表
CREATE TABLE system_setting_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建系统设置表
CREATE TABLE system_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category VARCHAR(100) NOT NULL,
    key VARCHAR(100) NOT NULL,
    value TEXT,
    default_value TEXT,
    value_type VARCHAR(20) DEFAULT 'string', -- string, integer, boolean, json
    description TEXT,
    is_required BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(category, key)
);

-- 创建系统设置历史表
CREATE TABLE system_setting_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    setting_id UUID NOT NULL REFERENCES system_settings(id) ON DELETE CASCADE,
    old_value TEXT,
    new_value TEXT,
    changed_by VARCHAR(100),
    change_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_contents_user_id ON contents(user_id);
CREATE INDEX idx_contents_created_at ON contents(created_at);
CREATE INDEX idx_contents_is_active ON contents(is_active);
CREATE INDEX idx_content_analytics_content_id ON content_analytics(content_id);
CREATE INDEX idx_content_analytics_access_time ON content_analytics(access_time);
CREATE INDEX idx_content_analytics_country ON content_analytics(country);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_api_key ON users(api_key);
CREATE INDEX idx_user_subscriptions_user_id ON user_subscriptions(user_id);
CREATE INDEX idx_user_subscriptions_status ON user_subscriptions(status);
CREATE INDEX idx_usage_statistics_user_month ON usage_statistics(user_id, month_year);
CREATE INDEX idx_system_settings_category_key ON system_settings(category, key);
CREATE INDEX idx_system_settings_active ON system_settings(is_active);

-- 插入默认管理员用户
INSERT INTO users (id, username, email, password, api_key, is_admin, is_active)
VALUES
    ('6d53e189-76e0-4c4a-b94d-4e942e25bf60', 'admin', 'admin@example.com', '$2a$10$jrRp8iAmLRbJigivn/5EseN1vjc7fHtbIY1YNmqGukYzWHbM7M99W', 'admin-api-key-' || substr(md5(random()::text), 1, 16), TRUE, TRUE),
    ('56ecc01e-9ffc-4515-8932-19912ce0805d', 'yingcai', 'yingcai@yingcai.com', '$2a$10$sBJC6zurK08jxGW1l8BQH.35v4fJ63TiYSJH0i3cWMMqUBPXnXeyC', 'yingcai-api-key-' || substr(md5(random()::text), 1, 16), TRUE, TRUE),
    ('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', 'testuser1', 'test1@example.com', '$2a$10$jrRp8iAmLRbJigivn/5EseN1vjc7fHtbIY1YNmqGukYzWHbM7M99W', 'test1-api-key-' || substr(md5(random()::text), 1, 16), FALSE, TRUE),
    ('bbbbbbbb-cccc-dddd-eeee-ffffffffffff', 'testuser2', 'test2@example.com', '$2a$10$jrRp8iAmLRbJigivn/5EseN1vjc7fHtbIY1YNmqGukYzWHbM7M99W', 'test2-api-key-' || substr(md5(random()::text), 1, 16), FALSE, TRUE);

-- 插入默认计划配置
INSERT INTO plan_configs (id, name, display_name, description, max_uploads_per_day, max_file_size_mb, max_storage_mb, retention_days, price_monthly, price_yearly, features, is_active, sort_order)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'free', '免费版', '适合个人用户的基础功能', 10, 10, 100, 7, 0.00, 0.00, '{"api_access": true, "analytics": false, "custom_domain": false}', TRUE, 1),
    ('22222222-2222-2222-2222-222222222222', 'basic', '基础版', '适合小团队的增强功能', 50, 50, 1000, 30, 9.99, 99.99, '{"api_access": true, "analytics": true, "custom_domain": false}', TRUE, 2),
    ('33333333-3333-3333-3333-333333333333', 'pro', '专业版', '适合企业用户的完整功能', 200, 200, 10000, 90, 29.99, 299.99, '{"api_access": true, "analytics": true, "custom_domain": true}', TRUE, 3);

-- 插入系统设置分类
INSERT INTO system_setting_categories (id, name, display_name, description, icon, sort_order, is_active)
VALUES
    ('c1111111-1111-1111-1111-111111111111', 'server', '服务器设置', '服务器相关配置，包括端口、主机等', 'bi-server', 1, TRUE),
    ('c2222222-2222-2222-2222-222222222222', 'database', '数据库设置', '数据库连接和配置参数', 'bi-database', 2, TRUE),
    ('c3333333-3333-3333-3333-333333333333', 'upload', '上传设置', '文件上传相关配置', 'bi-cloud-upload', 3, TRUE),
    ('c4444444-4444-4444-4444-444444444444', 'security', '安全设置', '安全相关配置，包括JWT、限流等', 'bi-shield-check', 4, TRUE),
    ('c5555555-5555-5555-5555-555555555555', 'geoip', '地理位置设置', 'GeoIP服务相关配置', 'bi-globe', 5, TRUE),
    ('c6666666-6666-6666-6666-666666666666', 'system', '系统设置', '系统级别的配置参数', 'bi-gear', 6, TRUE);

-- 插入用户订阅（为测试用户分配计划）
INSERT INTO user_subscriptions (id, user_id, plan_id, status, starts_at, expires_at, auto_renew)
VALUES
    ('s1111111-1111-1111-1111-111111111111', '6d53e189-76e0-4c4a-b94d-4e942e25bf60', '33333333-3333-3333-3333-333333333333', 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '1 year', TRUE),
    ('s2222222-2222-2222-2222-222222222222', '56ecc01e-9ffc-4515-8932-19912ce0805d', '33333333-3333-3333-3333-333333333333', 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '1 year', TRUE),
    ('s3333333-3333-3333-3333-333333333333', 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', '22222222-2222-2222-2222-222222222222', 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '1 month', FALSE),
    ('s4444444-4444-4444-4444-444444444444', 'bbbbbbbb-cccc-dddd-eeee-ffffffffffff', '11111111-1111-1111-1111-111111111111', 'active', CURRENT_TIMESTAMP, NULL, FALSE);

-- 插入测试内容
INSERT INTO contents (id, user_id, title, content, content_type, access_count, created_at)
VALUES
    ('c0000001-1111-1111-1111-111111111111', '6d53e189-76e0-4c4a-b94d-4e942e25bf60', '欢迎使用 AnyWebsites', '这是一个示例文本内容，展示了 AnyWebsites 的基本功能。您可以上传文本、图片、文档等各种类型的文件。', 'text/plain', 156, CURRENT_TIMESTAMP - INTERVAL '7 days'),
    ('c0000002-2222-2222-2222-222222222222', '56ecc01e-9ffc-4515-8932-19912ce0805d', 'API 使用指南', '# API 使用指南\n\n## 认证\n使用 Bearer Token 进行认证：\n```\nAuthorization: Bearer YOUR_API_KEY\n```\n\n## 上传文件\n```bash\ncurl -X POST https://api.anywebsites.com/api/content/upload \\\n  -H "Authorization: Bearer YOUR_API_KEY" \\\n  -F "file=@example.txt"\n```', 'text/markdown', 89, CURRENT_TIMESTAMP - INTERVAL '5 days'),
    ('c0000003-3333-3333-3333-333333333333', 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', '测试数据', '这是一些测试数据，用于演示系统功能。', 'text/plain', 23, CURRENT_TIMESTAMP - INTERVAL '3 days'),
    ('c0000004-4444-4444-4444-444444444444', 'bbbbbbbb-cccc-dddd-eeee-ffffffffffff', 'JSON 配置示例', '{"name": "example", "version": "1.0.0", "description": "示例配置文件"}', 'application/json', 45, CURRENT_TIMESTAMP - INTERVAL '2 days'),
    ('c0000005-5555-5555-5555-555555555555', '6d53e189-76e0-4c4a-b94d-4e942e25bf60', '系统公告', '系统将于本周末进行维护升级，预计停机时间为2小时。感谢您的理解与支持！', 'text/plain', 234, CURRENT_TIMESTAMP - INTERVAL '1 day');

-- 插入内容分析数据（模拟访问记录）
INSERT INTO content_analytics (id, content_id, ip_address, user_agent, referer, country, region, city, access_time)
VALUES
    -- 中国访问记录
    ('a0000001-1111-1111-1111-111111111111', 'c0000001-1111-1111-1111-111111111111', '192.168.1.100', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', 'https://google.com', 'China', 'Beijing', 'Beijing', CURRENT_TIMESTAMP - INTERVAL '6 days'),
    ('a0000002-2222-2222-2222-222222222222', 'c0000001-1111-1111-1111-111111111111', '192.168.1.101', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36', 'https://baidu.com', 'China', 'Shanghai', 'Shanghai', CURRENT_TIMESTAMP - INTERVAL '6 days'),
    ('a0000003-3333-3333-3333-333333333333', 'c0000002-2222-2222-2222-222222222222', '192.168.1.102', 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X)', 'https://github.com', 'China', 'Guangdong', 'Shenzhen', CURRENT_TIMESTAMP - INTERVAL '5 days'),

    -- 美国访问记录
    ('a0000004-4444-4444-4444-444444444444', 'c0000001-1111-1111-1111-111111111111', '203.0.113.1', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', 'https://google.com', 'United States', 'California', 'San Francisco', CURRENT_TIMESTAMP - INTERVAL '4 days'),
    ('a0000005-5555-5555-5555-555555555555', 'c0000002-2222-2222-2222-222222222222', '203.0.113.2', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36', 'https://stackoverflow.com', 'United States', 'New York', 'New York', CURRENT_TIMESTAMP - INTERVAL '4 days'),

    -- 日本访问记录
    ('a0000006-6666-6666-6666-666666666666', 'c0000003-3333-3333-3333-333333333333', '198.51.100.1', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', 'https://yahoo.co.jp', 'Japan', 'Tokyo', 'Tokyo', CURRENT_TIMESTAMP - INTERVAL '3 days'),

    -- 德国访问记录
    ('a0000007-7777-7777-7777-777777777777', 'c0000004-4444-4444-4444-444444444444', '198.51.100.2', 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36', 'https://google.de', 'Germany', 'Bavaria', 'Munich', CURRENT_TIMESTAMP - INTERVAL '2 days'),

    -- 英国访问记录
    ('a0000008-8888-8888-8888-888888888888', 'c0000005-5555-5555-5555-555555555555', '198.51.100.3', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', 'https://bbc.co.uk', 'United Kingdom', 'England', 'London', CURRENT_TIMESTAMP - INTERVAL '1 day'),

    -- 更多中国访问记录
    ('a0000009-9999-9999-9999-999999999999', 'c0000001-1111-1111-1111-111111111111', '192.168.1.103', 'Mozilla/5.0 (Android 11; Mobile) AppleWebKit/537.36', 'https://weibo.com', 'China', 'Zhejiang', 'Hangzhou', CURRENT_TIMESTAMP - INTERVAL '1 day'),
    ('a000000a-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'c0000002-2222-2222-2222-222222222222', '192.168.1.104', 'Mozilla/5.0 (iPad; CPU OS 14_7_1 like Mac OS X)', 'https://zhihu.com', 'China', 'Jiangsu', 'Nanjing', CURRENT_TIMESTAMP - INTERVAL '12 hours');

-- 插入使用统计数据
INSERT INTO usage_statistics (id, user_id, month_year, uploads_count, storage_used_mb, api_calls_count, bandwidth_used_mb)
VALUES
    ('u0000001-1111-1111-1111-111111111111', '6d53e189-76e0-4c4a-b94d-4e942e25bf60', '2025-06', 45, 1250, 234, 5600),
    ('u0000002-2222-2222-2222-222222222222', '56ecc01e-9ffc-4515-8932-19912ce0805d', '2025-06', 32, 890, 156, 3400),
    ('u0000003-3333-3333-3333-333333333333', 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', '2025-06', 18, 456, 89, 1200),
    ('u0000004-4444-4444-4444-444444444444', 'bbbbbbbb-cccc-dddd-eeee-ffffffffffff', '2025-06', 8, 123, 45, 567),
    ('u0000005-5555-5555-5555-555555555555', '6d53e189-76e0-4c4a-b94d-4e942e25bf60', '2025-05', 38, 1100, 198, 4800),
    ('u0000006-6666-6666-6666-666666666666', '56ecc01e-9ffc-4515-8932-19912ce0805d', '2025-05', 28, 750, 134, 2900),
    ('u0000007-7777-7777-7777-777777777777', 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', '2025-05', 15, 380, 67, 980),
    ('u0000008-8888-8888-8888-888888888888', 'bbbbbbbb-cccc-dddd-eeee-ffffffffffff', '2025-05', 5, 89, 23, 234);

-- 插入计划升级历史
INSERT INTO plan_upgrade_history (id, user_id, from_plan_id, to_plan_id, upgrade_type, reason, effective_date)
VALUES
    ('h0000001-1111-1111-1111-111111111111', 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'upgrade', '需要更多存储空间和分析功能', CURRENT_TIMESTAMP - INTERVAL '30 days'),
    ('h0000002-2222-2222-2222-222222222222', '6d53e189-76e0-4c4a-b94d-4e942e25bf60', '22222222-2222-2222-2222-222222222222', '33333333-3333-3333-3333-333333333333', 'upgrade', '管理员账户升级到专业版', CURRENT_TIMESTAMP - INTERVAL '60 days'),
    ('h0000003-3333-3333-3333-333333333333', '56ecc01e-9ffc-4515-8932-19912ce0805d', '11111111-1111-1111-1111-111111111111', '33333333-3333-3333-3333-333333333333', 'upgrade', '直接升级到专业版以获得完整功能', CURRENT_TIMESTAMP - INTERVAL '90 days');

-- 插入一些基础系统设置
INSERT INTO system_settings (id, category, key, value, default_value, value_type, description, is_required, is_active)
VALUES
    ('s0000001-1111-1111-1111-111111111111', 'server', 'host', '0.0.0.0', '0.0.0.0', 'string', '服务器监听地址', TRUE, TRUE),
    ('s0000002-2222-2222-2222-222222222222', 'server', 'port', '8085', '8085', 'integer', '服务器监听端口', TRUE, TRUE),
    ('s0000003-3333-3333-3333-333333333333', 'upload', 'max_file_size', '100', '100', 'integer', '最大文件大小（MB）', TRUE, TRUE),
    ('s0000004-4444-4444-4444-444444444444', 'upload', 'allowed_types', 'text/*,image/*,application/json,application/pdf', 'text/*,image/*', 'string', '允许的文件类型', TRUE, TRUE),
    ('s0000005-5555-5555-5555-555555555555', 'security', 'jwt_secret', 'your-secret-key-here', 'change-me', 'string', 'JWT 密钥', TRUE, TRUE),
    ('s0000006-6666-6666-6666-666666666666', 'security', 'rate_limit', '100', '100', 'integer', '每分钟请求限制', TRUE, TRUE),
    ('s0000007-7777-7777-7777-777777777777', 'geoip', 'enabled', 'true', 'false', 'boolean', '启用地理位置服务', FALSE, TRUE),
    ('s0000008-8888-8888-8888-888888888888', 'geoip', 'database_path', '/app/data/GeoLite2-City.mmdb', '', 'string', 'GeoIP 数据库路径', FALSE, TRUE),
    ('s0000009-9999-9999-9999-999999999999', 'system', 'site_name', 'AnyWebsites', 'AnyWebsites', 'string', '网站名称', FALSE, TRUE),
    ('s000000a-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'system', 'maintenance_mode', 'false', 'false', 'boolean', '维护模式', FALSE, TRUE);