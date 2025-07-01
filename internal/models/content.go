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
	Title       string     `json:"title" gorm:"size:255"`
	Description string     `json:"description" gorm:"size:500"`
	Content     string     `json:"content" gorm:"type:text;not null;column:content"`
	ContentType string     `json:"content_type" gorm:"size:50;default:'text/html'"`
	FilePath    string     `json:"file_path" gorm:"size:500"`
	FileSize    int64      `json:"file_size" gorm:"default:0"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	AccessCount int        `json:"access_count" gorm:"default:0"`
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
func (c *Content) CanAccess() bool {
	if !c.IsActive {
		return false
	}
	if c.IsExpired() {
		return false
	}
	return true
}
