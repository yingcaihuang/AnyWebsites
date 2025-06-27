package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PlanType 计划类型枚举
type PlanType string

const (
	PlanCommunity  PlanType = "community"
	PlanDeveloper  PlanType = "developer"
	PlanPro        PlanType = "pro"
	PlanMax        PlanType = "max"
	PlanEnterprise PlanType = "enterprise"
)

// SubscriptionStatus 订阅状态枚举
type SubscriptionStatus string

const (
	StatusActive    SubscriptionStatus = "active"
	StatusExpired   SubscriptionStatus = "expired"
	StatusCancelled SubscriptionStatus = "cancelled"
	StatusPending   SubscriptionStatus = "pending"
	StatusSuspended SubscriptionStatus = "suspended"
)

// PlanConfig 计划配置模型
type PlanConfig struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Type                 PlanType  `gorm:"type:varchar(20);unique;not null" json:"type"`
	Name                 string    `gorm:"type:varchar(100);not null" json:"name"`
	Price                float64   `gorm:"type:decimal(10,2);not null;default:0" json:"price"`
	Currency             string    `gorm:"type:varchar(3);not null;default:'USD'" json:"currency"`
	ArticleRetentionDays int       `gorm:"not null" json:"article_retention_days"`
	MonthlyUploadLimit   int       `gorm:"not null" json:"monthly_upload_limit"`
	StorageLimitMB       int64     `gorm:"not null" json:"storage_limit_mb"`
	APIRateLimitPerHour  int       `gorm:"not null" json:"api_rate_limit_per_hour"`
	Features             string    `gorm:"type:text" json:"features"`
	IsActive             bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// UserSubscription 用户订阅模型
type UserSubscription struct {
	ID            uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID          `gorm:"type:uuid;not null" json:"user_id"`
	PlanType      PlanType           `gorm:"type:varchar(20);not null" json:"plan_type"`
	Status        SubscriptionStatus `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	StartedAt     time.Time          `gorm:"not null;default:now()" json:"started_at"`
	ExpiresAt     *time.Time         `json:"expires_at"`
	AutoRenew     bool               `gorm:"not null;default:false" json:"auto_renew"`
	PaymentMethod string             `gorm:"type:varchar(50)" json:"payment_method"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`

	// 关联 - 暂时移除以避免关联问题
	// PlanConfig PlanConfig `gorm:"foreignKey:PlanType;references:Type" json:"plan_config,omitempty"`
}

// UsageStatistics 使用量统计模型
type UsageStatistics struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	MonthYear        string    `gorm:"type:varchar(7);not null" json:"month_year"` // 格式: 2024-01
	ArticlesUploaded int       `gorm:"not null;default:0" json:"articles_uploaded"`
	StorageUsedMB    int64     `gorm:"not null;default:0" json:"storage_used_mb"`
	APICallsMade     int       `gorm:"not null;default:0" json:"api_calls_made"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// 关联 - 移除以避免循环导入
}

// PlanUpgradeHistory 计划升级历史
type PlanUpgradeHistory struct {
	ID           uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID          `gorm:"type:uuid;not null" json:"user_id"`
	FromPlan     PlanType           `gorm:"type:varchar(20)" json:"from_plan"`
	ToPlan       PlanType           `gorm:"type:varchar(20);not null" json:"to_plan"`
	ChangeReason string             `gorm:"type:varchar(100)" json:"change_reason"`
	EffectiveAt  time.Time          `gorm:"not null" json:"effective_at"`
	Status       SubscriptionStatus `gorm:"type:varchar(20);not null" json:"status"`
	CreatedAt    time.Time          `json:"created_at"`

	// 关联 - 移除以避免循环导入
}

// BeforeCreate 创建前钩子
func (pc *PlanConfig) BeforeCreate(tx *gorm.DB) error {
	if pc.ID == uuid.Nil {
		pc.ID = uuid.New()
	}
	return nil
}

func (us *UserSubscription) BeforeCreate(tx *gorm.DB) error {
	if us.ID == uuid.Nil {
		us.ID = uuid.New()
	}
	return nil
}

func (usage *UsageStatistics) BeforeCreate(tx *gorm.DB) error {
	if usage.ID == uuid.Nil {
		usage.ID = uuid.New()
	}
	return nil
}

func (puh *PlanUpgradeHistory) BeforeCreate(tx *gorm.DB) error {
	if puh.ID == uuid.Nil {
		puh.ID = uuid.New()
	}
	return nil
}

// GetDefaultPlanConfigs 获取默认计划配置
func GetDefaultPlanConfigs() []PlanConfig {
	return []PlanConfig{
		{
			Type:                 PlanCommunity,
			Name:                 "Community Plan",
			Price:                0,
			Currency:             "USD",
			ArticleRetentionDays: 7,
			MonthlyUploadLimit:   50,
			StorageLimitMB:       100,
			APIRateLimitPerHour:  100,
			Features:             `["50 articles per month","7 days retention","100MB storage","Public articles only","Basic statistics","Community support"]`,
			IsActive:             true,
		},
		{
			Type:                 PlanDeveloper,
			Name:                 "Developer Plan",
			Price:                50.00,
			Currency:             "USD",
			ArticleRetentionDays: 30,
			MonthlyUploadLimit:   600,
			StorageLimitMB:       1024,
			APIRateLimitPerHour:  1000,
			Features:             `["600 articles per month","30 days retention","1GB storage","Private articles with access codes","Basic custom domain","Detailed analytics","Email support","Team collaboration"]`,
			IsActive:             true,
		},
		{
			Type:                 PlanPro,
			Name:                 "Pro Plan",
			Price:                100.00,
			Currency:             "USD",
			ArticleRetentionDays: 90,
			MonthlyUploadLimit:   1500,
			StorageLimitMB:       5120,
			APIRateLimitPerHour:  5000,
			Features:             `["1500 articles per month","90 days retention","5GB storage","Advanced custom domains","White-label solution","Advanced analytics and reports","Priority support","Advanced team management","Custom themes"]`,
			IsActive:             true,
		},
		{
			Type:                 PlanMax,
			Name:                 "Max Plan",
			Price:                250.00,
			Currency:             "USD",
			ArticleRetentionDays: 365,
			MonthlyUploadLimit:   4500,
			StorageLimitMB:       20480,
			APIRateLimitPerHour:  20000,
			Features:             `["4500 articles per month","365 days retention","20GB storage","Unlimited custom domains","Complete white-label","Real-time monitoring","24/7 dedicated support","Enterprise security","API priority"]`,
			IsActive:             true,
		},
		{
			Type:                 PlanEnterprise,
			Name:                 "Enterprise Plan",
			Price:                0, // 联系销售
			Currency:             "USD",
			ArticleRetentionDays: -1, // 无限制
			MonthlyUploadLimit:   -1, // 无限制
			StorageLimitMB:       -1, // 无限制
			APIRateLimitPerHour:  -1, // 无限制
			Features:             `["Unlimited articles","Unlimited retention","Unlimited storage","Custom solutions","Dedicated servers","SSO integration","Compliance support","Dedicated account manager","SLA guarantee"]`,
			IsActive:             true,
		},
	}
}

// IsUnlimited 检查是否为无限制
func (pc *PlanConfig) IsUnlimited() bool {
	return pc.Type == PlanEnterprise
}

// IsExpired 检查订阅是否过期
func (us *UserSubscription) IsExpired() bool {
	if us.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*us.ExpiresAt)
}

// IsActive 检查订阅是否激活
func (us *UserSubscription) IsActive() bool {
	return us.Status == StatusActive && !us.IsExpired()
}

// GetCurrentMonthYear 获取当前月份年份字符串
func GetCurrentMonthYear() string {
	return time.Now().Format("2006-01")
}

// TableName 指定表名
func (PlanConfig) TableName() string {
	return "plan_configs"
}

func (UserSubscription) TableName() string {
	return "user_subscriptions"
}

func (UsageStatistics) TableName() string {
	return "usage_statistics"
}

func (PlanUpgradeHistory) TableName() string {
	return "plan_upgrade_histories"
}
