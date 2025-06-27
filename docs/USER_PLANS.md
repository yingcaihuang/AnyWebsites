# 用户等级计划系统设计

## 📋 概述

AnyWebsites 用户等级计划系统旨在为不同需求的用户提供差异化的服务，通过等级限制和功能差异实现商业化运营。

## 🎯 等级计划定义

### 1. Community Plan (免费版)
- **价格**: 免费
- **文章保存期限**: 7天
- **每月上传限制**: 50篇文章
- **存储空间**: 100MB
- **API调用频率**: 100次/小时
- **功能限制**:
  - 仅支持公开文章
  - 不支持自定义域名
  - 基础统计功能
  - 社区支持

### 2. Developer Plan (开发者版)
- **价格**: $50.00/月
- **文章保存期限**: 30天
- **每月上传限制**: 600篇文章
- **存储空间**: 1GB
- **API调用频率**: 1000次/小时
- **功能特性**:
  - 支持私有文章和访问码
  - 基础自定义域名
  - 详细统计和分析
  - 邮件支持
  - 团队协作功能

### 3. Pro Plan (专业版)
- **价格**: $100.00/月
- **文章保存期限**: 90天
- **每月上传限制**: 1500篇文章
- **存储空间**: 5GB
- **API调用频率**: 5000次/小时
- **功能特性**:
  - 高级自定义域名
  - 白标解决方案
  - 高级统计和报告
  - 优先技术支持
  - 高级团队管理
  - 自定义主题

### 4. Max Plan (最大版)
- **价格**: $250.00/月
- **文章保存期限**: 365天
- **每月上传限制**: 4500篇文章
- **存储空间**: 20GB
- **API调用频率**: 20000次/小时
- **功能特性**:
  - 无限自定义域名
  - 完全白标
  - 实时统计和监控
  - 24/7专属支持
  - 企业级安全
  - API优先级

### 5. Enterprise Plan (企业版)
- **价格**: 联系销售
- **文章保存期限**: 无限制
- **每月上传限制**: 无限制
- **存储空间**: 无限制
- **API调用频率**: 无限制
- **功能特性**:
  - 定制化解决方案
  - 专属服务器
  - SSO集成
  - 合规性支持
  - 专属客户经理
  - SLA保证

## 🔧 技术实现规范

### 等级枚举定义
```go
type PlanType string

const (
    PlanCommunity  PlanType = "community"
    PlanDeveloper  PlanType = "developer"
    PlanPro        PlanType = "pro"
    PlanMax        PlanType = "max"
    PlanEnterprise PlanType = "enterprise"
)
```

### 等级配置结构
```go
type PlanConfig struct {
    Type                PlanType      `json:"type"`
    Name                string        `json:"name"`
    Price               float64       `json:"price"`
    Currency            string        `json:"currency"`
    ArticleRetentionDays int          `json:"article_retention_days"`
    MonthlyUploadLimit  int           `json:"monthly_upload_limit"`
    StorageLimit        int64         `json:"storage_limit_mb"`
    APIRateLimit        int           `json:"api_rate_limit_per_hour"`
    Features            []string      `json:"features"`
    IsActive            bool          `json:"is_active"`
}
```

### 用户订阅状态
```go
type SubscriptionStatus string

const (
    StatusActive    SubscriptionStatus = "active"
    StatusExpired   SubscriptionStatus = "expired"
    StatusCancelled SubscriptionStatus = "cancelled"
    StatusPending   SubscriptionStatus = "pending"
)
```

## 📊 数据库设计

### 1. 计划配置表 (plan_configs)
```sql
CREATE TABLE plan_configs (
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
```

### 2. 用户订阅表 (user_subscriptions)
```sql
CREATE TABLE user_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    plan_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    auto_renew BOOLEAN NOT NULL DEFAULT false,
    payment_method VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (plan_type) REFERENCES plan_configs(type)
);
```

### 3. 使用量统计表 (usage_statistics)
```sql
CREATE TABLE usage_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    month_year VARCHAR(7) NOT NULL, -- 格式: 2024-01
    articles_uploaded INTEGER NOT NULL DEFAULT 0,
    storage_used_mb BIGINT NOT NULL DEFAULT 0,
    api_calls_made INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(user_id, month_year)
);
```

## 🚀 业务逻辑规则

### 1. 文章过期规则
- 文章创建时根据用户当前等级设置过期时间
- 用户升级时，现有文章过期时间延长
- 用户降级时，现有文章过期时间缩短（但不会立即删除）
- 过期文章进入软删除状态，30天后硬删除

### 2. 限制检查规则
- 上传文章前检查月度限制
- API调用前检查频率限制
- 存储使用前检查空间限制
- 超限时返回相应错误码和升级提示

### 3. 等级变更规则
- 升级立即生效
- 降级在当前计费周期结束后生效
- 取消订阅后降级为免费版
- 支持等级暂停和恢复

## 📈 监控和分析

### 1. 使用量监控
- 实时统计用户使用量
- 接近限制时发送通知
- 生成使用量报告

### 2. 收入分析
- 按等级统计收入
- 用户转化率分析
- 流失用户分析

### 3. 功能使用分析
- 各等级功能使用率
- 用户行为分析
- 产品优化建议

## 🔒 安全考虑

### 1. 权限控制
- 基于等级的功能访问控制
- API密钥等级标识
- 管理员权限分离

### 2. 防滥用机制
- 频率限制
- 异常检测
- 自动封禁机制

### 3. 数据保护
- 用户数据隔离
- 敏感信息加密
- 审计日志记录

---

*设计文档版本: v1.0*  
*最后更新: 2025-06-24*
