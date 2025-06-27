-- 创建计划配置表
CREATE TABLE IF NOT EXISTS plan_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(10,2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    article_retention_days INTEGER NOT NULL,
    monthly_upload_limit INTEGER NOT NULL,
    storage_limit_mb BIGINT NOT NULL,
    api_rate_limit_per_hour INTEGER NOT NULL,
    features JSONB,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建用户订阅表
CREATE TABLE IF NOT EXISTS user_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    auto_renew BOOLEAN NOT NULL DEFAULT false,
    payment_method VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (plan_type) REFERENCES plan_configs(type) ON DELETE RESTRICT
);

-- 创建使用量统计表
CREATE TABLE IF NOT EXISTS usage_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    month_year VARCHAR(7) NOT NULL, -- 格式: 2024-01
    articles_uploaded INTEGER NOT NULL DEFAULT 0,
    storage_used_mb BIGINT NOT NULL DEFAULT 0,
    api_calls_made INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, month_year)
);

-- 创建计划升级历史表
CREATE TABLE IF NOT EXISTS plan_upgrade_histories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_plan VARCHAR(20),
    to_plan VARCHAR(20) NOT NULL,
    change_reason VARCHAR(100),
    effective_at TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (from_plan) REFERENCES plan_configs(type) ON DELETE SET NULL,
    FOREIGN KEY (to_plan) REFERENCES plan_configs(type) ON DELETE RESTRICT
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_user_id ON user_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_plan_type ON user_subscriptions(plan_type);
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_status ON user_subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_expires_at ON user_subscriptions(expires_at);

CREATE INDEX IF NOT EXISTS idx_usage_statistics_user_id ON usage_statistics(user_id);
CREATE INDEX IF NOT EXISTS idx_usage_statistics_month_year ON usage_statistics(month_year);

CREATE INDEX IF NOT EXISTS idx_plan_upgrade_histories_user_id ON plan_upgrade_histories(user_id);
CREATE INDEX IF NOT EXISTS idx_plan_upgrade_histories_effective_at ON plan_upgrade_histories(effective_at);

-- 添加触发器更新 updated_at 字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_plan_configs_updated_at BEFORE UPDATE ON plan_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_subscriptions_updated_at BEFORE UPDATE ON user_subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_usage_statistics_updated_at BEFORE UPDATE ON usage_statistics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 插入默认计划配置
INSERT INTO plan_configs (type, name, price, currency, article_retention_days, monthly_upload_limit, storage_limit_mb, api_rate_limit_per_hour, features, is_active) VALUES
('community', 'Community Plan', 0, 'USD', 7, 50, 100, 100, 
 '["50 articles per month", "7 days retention", "100MB storage", "Public articles only", "Basic statistics", "Community support"]', true),

('developer', 'Developer Plan', 50.00, 'USD', 30, 600, 1024, 1000,
 '["600 articles per month", "30 days retention", "1GB storage", "Private articles with access codes", "Basic custom domain", "Detailed analytics", "Email support", "Team collaboration"]', true),

('pro', 'Pro Plan', 100.00, 'USD', 90, 1500, 5120, 5000,
 '["1500 articles per month", "90 days retention", "5GB storage", "Advanced custom domains", "White-label solution", "Advanced analytics and reports", "Priority support", "Advanced team management", "Custom themes"]', true),

('max', 'Max Plan', 250.00, 'USD', 365, 4500, 20480, 20000,
 '["4500 articles per month", "365 days retention", "20GB storage", "Unlimited custom domains", "Complete white-label", "Real-time monitoring", "24/7 dedicated support", "Enterprise security", "API priority"]', true),

('enterprise', 'Enterprise Plan', 0, 'USD', -1, -1, -1, -1,
 '["Unlimited articles", "Unlimited retention", "Unlimited storage", "Custom solutions", "Dedicated servers", "SSO integration", "Compliance support", "Dedicated account manager", "SLA guarantee"]', true)

ON CONFLICT (type) DO UPDATE SET
    name = EXCLUDED.name,
    price = EXCLUDED.price,
    currency = EXCLUDED.currency,
    article_retention_days = EXCLUDED.article_retention_days,
    monthly_upload_limit = EXCLUDED.monthly_upload_limit,
    storage_limit_mb = EXCLUDED.storage_limit_mb,
    api_rate_limit_per_hour = EXCLUDED.api_rate_limit_per_hour,
    features = EXCLUDED.features,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- 为现有用户创建默认的社区版订阅
INSERT INTO user_subscriptions (user_id, plan_type, status, started_at)
SELECT id, 'community', 'active', NOW()
FROM users
WHERE id NOT IN (SELECT user_id FROM user_subscriptions)
ON CONFLICT DO NOTHING;

-- 添加注释
COMMENT ON TABLE plan_configs IS '用户计划配置表';
COMMENT ON TABLE user_subscriptions IS '用户订阅表';
COMMENT ON TABLE usage_statistics IS '用户使用量统计表';
COMMENT ON TABLE plan_upgrade_histories IS '计划升级历史表';

COMMENT ON COLUMN plan_configs.article_retention_days IS '文章保存天数，-1表示无限制';
COMMENT ON COLUMN plan_configs.monthly_upload_limit IS '每月上传限制，-1表示无限制';
COMMENT ON COLUMN plan_configs.storage_limit_mb IS '存储限制(MB)，-1表示无限制';
COMMENT ON COLUMN plan_configs.api_rate_limit_per_hour IS 'API频率限制(每小时)，-1表示无限制';

COMMENT ON COLUMN user_subscriptions.expires_at IS '订阅过期时间，NULL表示永不过期';
COMMENT ON COLUMN user_subscriptions.auto_renew IS '是否自动续费';

COMMENT ON COLUMN usage_statistics.month_year IS '统计月份，格式：YYYY-MM';
COMMENT ON COLUMN usage_statistics.articles_uploaded IS '当月上传文章数';
COMMENT ON COLUMN usage_statistics.storage_used_mb IS '当月使用存储(MB)';
COMMENT ON COLUMN usage_statistics.api_calls_made IS '当月API调用次数';
