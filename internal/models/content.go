package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Content HTML 内容模型
type Content struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Title       string     `json:"title" gorm:"size:200"`
	Description string     `json:"description" gorm:"size:500"`
	HTMLContent string     `json:"html_content" gorm:"type:text;not null"`
	AccessCode  string     `json:"access_code,omitempty" gorm:"size:64"` // 加密访问码
	IsPublic    bool       `json:"is_public" gorm:"default:true"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	ViewCount   int64      `json:"view_count" gorm:"default:0"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// 关联关系
	User      User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Analytics []ContentAnalytics `json:"analytics,omitempty" gorm:"foreignKey:ContentID"`
}

// BeforeCreate 在创建内容前生成 UUID
func (c *Content) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// IsExpired 检查内容是否已过期
func (c *Content) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiresAt)
}

// CanAccess 检查是否可以访问内容
func (c *Content) CanAccess(accessCode string) bool {
	if !c.IsActive {
		return false
	}
	if c.IsExpired() {
		return false
	}
	if c.IsPublic {
		return true
	}
	return c.AccessCode == accessCode
}
