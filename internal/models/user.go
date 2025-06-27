package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null;size:100"`
	Password  string    `json:"-" gorm:"not null;size:255"` // 不在 JSON 中返回密码
	APIKey    string    `json:"api_key" gorm:"uniqueIndex;not null;size:64"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	IsAdmin   bool      `json:"is_admin" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联关系
	Contents     []Content         `json:"contents,omitempty" gorm:"foreignKey:UserID"`
	Subscription *UserSubscription `json:"subscription,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate 在创建用户前生成 UUID 和 API Key
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.APIKey == "" {
		u.APIKey = generateAPIKey()
	}
	return nil
}

// generateAPIKey 生成 API 密钥
func generateAPIKey() string {
	return uuid.New().String()[:32] // 只取前32个字符
}
