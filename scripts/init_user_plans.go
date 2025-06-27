package main

import (
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 简化的模型定义
type User struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	Username string    `gorm:"type:varchar(50);unique;not null"`
	Email    string    `gorm:"type:varchar(100);unique;not null"`
}

type PlanConfig struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type                 string    `gorm:"type:varchar(20);unique;not null"`
	Name                 string    `gorm:"type:varchar(100);not null"`
	Price                float64   `gorm:"type:decimal(10,2);not null;default:0"`
	Currency             string    `gorm:"type:varchar(3);not null;default:'USD'"`
	ArticleRetentionDays int       `gorm:"not null"`
	MonthlyUploadLimit   int       `gorm:"not null"`
	StorageLimitMB       int64     `gorm:"not null"`
	APIRateLimitPerHour  int       `gorm:"not null"`
	Features             string    `gorm:"type:text"`
	IsActive             bool      `gorm:"not null;default:true"`
}

type UserSubscription struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        uuid.UUID `gorm:"type:uuid;not null"`
	PlanType      string    `gorm:"type:varchar(20);not null"`
	Status        string    `gorm:"type:varchar(20);not null;default:'active'"`
	StartedAt     string    `gorm:"not null;default:now()"`
	ExpiresAt     *string   `json:"expires_at"`
	AutoRenew     bool      `gorm:"not null;default:false"`
	PaymentMethod string    `gorm:"type:varchar(50)"`
}

func main() {
	// 数据库连接
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// 默认使用 Docker 容器内的连接
		dsn = "host=anywebsites-postgres-1 user=anywebsites password=anywebsites dbname=anywebsites port=5432 sslmode=disable TimeZone=UTC"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 1. 初始化计划配置
	fmt.Println("🔧 初始化计划配置...")
	if err := initPlanConfigs(db); err != nil {
		log.Fatal("Failed to init plan configs:", err)
	}

	// 2. 为现有用户创建默认订阅
	fmt.Println("👥 为现有用户创建默认订阅...")
	if err := createDefaultSubscriptions(db); err != nil {
		log.Fatal("Failed to create default subscriptions:", err)
	}

	fmt.Println("✅ 用户计划初始化完成！")
}

func initPlanConfigs(db *gorm.DB) error {
	configs := []PlanConfig{
		{
			Type:                 "community",
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
			Type:                 "developer",
			Name:                 "Developer Plan",
			Price:                50.00,
			Currency:             "USD",
			ArticleRetentionDays: 30,
			MonthlyUploadLimit:   600,
			StorageLimitMB:       1024,
			APIRateLimitPerHour:  1000,
			Features:             `["600 articles per month","30 days retention","1GB storage","Private articles","Access codes","Basic custom domain","Detailed analytics","Email support","Team collaboration"]`,
			IsActive:             true,
		},
		{
			Type:                 "pro",
			Name:                 "Pro Plan",
			Price:                100.00,
			Currency:             "USD",
			ArticleRetentionDays: 90,
			MonthlyUploadLimit:   1500,
			StorageLimitMB:       5120,
			APIRateLimitPerHour:  5000,
			Features:             `["1500 articles per month","90 days retention","5GB storage","Advanced custom domain","White-label solution","Advanced analytics","Priority support","Advanced team management","Custom themes"]`,
			IsActive:             true,
		},
		{
			Type:                 "max",
			Name:                 "Max Plan",
			Price:                250.00,
			Currency:             "USD",
			ArticleRetentionDays: 365,
			MonthlyUploadLimit:   5000,
			StorageLimitMB:       20480,
			APIRateLimitPerHour:  20000,
			Features:             `["5000 articles per month","1 year retention","20GB storage","Premium custom domain","Full white-label","Premium analytics","24/7 priority support","Enterprise team features","Advanced customization","API access"]`,
			IsActive:             true,
		},
		{
			Type:                 "enterprise",
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

	for _, config := range configs {
		var existing PlanConfig
		result := db.Where("type = ?", config.Type).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&config).Error; err != nil {
				return fmt.Errorf("failed to create plan config %s: %w", config.Type, err)
			}
			fmt.Printf("✅ 创建计划配置: %s\n", config.Name)
		} else {
			fmt.Printf("⏭️  计划配置已存在: %s\n", config.Name)
		}
	}

	return nil
}

func createDefaultSubscriptions(db *gorm.DB) error {
	// 获取所有没有订阅的用户
	var users []User
	if err := db.Raw(`
		SELECT u.id, u.username, u.email 
		FROM users u 
		LEFT JOIN user_subscriptions us ON u.id = us.user_id 
		WHERE us.id IS NULL
	`).Scan(&users).Error; err != nil {
		return fmt.Errorf("failed to get users without subscriptions: %w", err)
	}

	for _, user := range users {
		subscription := UserSubscription{
			UserID:        user.ID,
			PlanType:      "community",
			Status:        "active",
			StartedAt:     "now()",
			AutoRenew:     false,
			PaymentMethod: "",
		}

		if err := db.Create(&subscription).Error; err != nil {
			return fmt.Errorf("failed to create subscription for user %s: %w", user.Username, err)
		}

		fmt.Printf("✅ 为用户 %s 创建默认订阅\n", user.Username)
	}

	if len(users) == 0 {
		fmt.Println("⏭️  所有用户都已有订阅")
	}

	return nil
}
