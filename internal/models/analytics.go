package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContentAnalytics 内容访问统计模型
type ContentAnalytics struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ContentID uuid.UUID `json:"content_id" gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`

	// 访问信息
	IPAddress string `json:"ip_address" gorm:"size:45;index"` // 支持 IPv6
	UserAgent string `json:"user_agent" gorm:"size:500"`
	Referer   string `json:"referer" gorm:"size:500"`

	// 地理位置信息
	Country   string  `json:"country" gorm:"size:100"`
	Region    string  `json:"region" gorm:"size:100"`
	City      string  `json:"city" gorm:"size:100"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	// 时间信息
	AccessTime time.Time `json:"access_time" gorm:"index"`
	CreatedAt  time.Time `json:"created_at"`
}

// BeforeCreate 在创建分析记录前生成 UUID
func (ca *ContentAnalytics) BeforeCreate(tx *gorm.DB) error {
	if ca.ID == uuid.Nil {
		ca.ID = uuid.New()
	}
	if ca.AccessTime.IsZero() {
		ca.AccessTime = time.Now()
	}
	return nil
}

// TrafficStats 流量统计结构
type TrafficStats struct {
	Date      string `json:"date"`
	Views     int64  `json:"views"`
	UniqueIPs int64  `json:"unique_ips"`
}

// GeoStats 地理位置统计结构
type GeoStats struct {
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
	Count   int64  `json:"count"`
}

// CountryStats 国家分布统计结构
type CountryStats struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}

// RefererStats 来源统计结构
type RefererStats struct {
	Referer string `json:"referer"`
	Count   int64  `json:"count"`
}

// OverviewStats 总览统计结构
type OverviewStats struct {
	TotalViews     int64 `json:"total_views"`
	TotalContents  int64 `json:"total_contents"`
	ActiveContents int64 `json:"active_contents"`
	TodayViews     int64 `json:"today_views"`
	UniqueVisitors int64 `json:"unique_visitors"`
}
