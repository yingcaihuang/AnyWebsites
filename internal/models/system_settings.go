package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SystemSetting 系统设置模型
type SystemSetting struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Category    string    `gorm:"not null;index" json:"category"` // 设置分类：server, database, upload, security, etc.
	Key         string    `gorm:"not null;index" json:"key"`      // 设置键名
	Value       string    `gorm:"type:text" json:"value"`         // 设置值（JSON格式）
	ValueType   string    `gorm:"not null" json:"value_type"`     // 值类型：string, int, bool, json
	Description string    `gorm:"type:text" json:"description"`   // 设置描述
	IsActive    bool      `gorm:"default:true" json:"is_active"`  // 是否启用
	IsSystem    bool      `gorm:"default:false" json:"is_system"` // 是否为系统内置设置
	Version     int       `gorm:"default:1" json:"version"`       // 版本号
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   uuid.UUID `gorm:"type:uuid" json:"created_by"` // 创建者ID
	UpdatedBy   uuid.UUID `gorm:"type:uuid" json:"updated_by"` // 更新者ID

	// 关联
	Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updater User `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
}

// SystemSettingHistory 系统设置历史记录
type SystemSettingHistory struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SettingID  uuid.UUID `gorm:"type:uuid;not null;index" json:"setting_id"`
	OldValue   string    `gorm:"type:text" json:"old_value"`
	NewValue   string    `gorm:"type:text" json:"new_value"`
	ChangeType string    `gorm:"not null" json:"change_type"` // create, update, delete
	Reason     string    `gorm:"type:text" json:"reason"`     // 修改原因
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  uuid.UUID `gorm:"type:uuid" json:"created_by"`

	// 关联
	Setting SystemSetting `gorm:"foreignKey:SettingID" json:"setting,omitempty"`
	Creator User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// SystemSettingCategory 系统设置分类
type SystemSettingCategory struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"unique;not null" json:"name"`   // 分类名称
	DisplayName string    `gorm:"not null" json:"display_name"`  // 显示名称
	Description string    `gorm:"type:text" json:"description"`  // 分类描述
	Icon        string    `json:"icon"`                          // 图标
	SortOrder   int       `gorm:"default:0" json:"sort_order"`   // 排序
	IsActive    bool      `gorm:"default:true" json:"is_active"` // 是否启用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联
	Settings []SystemSetting `gorm:"foreignKey:Category;references:Name" json:"settings,omitempty"`
}

// TableName 指定表名
func (SystemSetting) TableName() string {
	return "system_settings"
}

func (SystemSettingHistory) TableName() string {
	return "system_setting_histories"
}

func (SystemSettingCategory) TableName() string {
	return "system_setting_categories"
}

// BeforeCreate 创建前钩子
func (s *SystemSetting) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (h *SystemSettingHistory) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}

func (c *SystemSettingCategory) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// GetStringValue 获取字符串值
func (s *SystemSetting) GetStringValue() string {
	return s.Value
}

// GetIntValue 获取整数值
func (s *SystemSetting) GetIntValue() (int, error) {
	var value int
	err := json.Unmarshal([]byte(s.Value), &value)
	return value, err
}

// GetBoolValue 获取布尔值
func (s *SystemSetting) GetBoolValue() (bool, error) {
	var value bool
	err := json.Unmarshal([]byte(s.Value), &value)
	return value, err
}

// GetJSONValue 获取JSON值
func (s *SystemSetting) GetJSONValue(v interface{}) error {
	return json.Unmarshal([]byte(s.Value), v)
}

// SetValue 设置值
func (s *SystemSetting) SetValue(value interface{}) error {
	switch v := value.(type) {
	case string:
		s.Value = v
		s.ValueType = "string"
	case int, int32, int64:
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		s.Value = string(data)
		s.ValueType = "int"
	case bool:
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		s.Value = string(data)
		s.ValueType = "bool"
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		s.Value = string(data)
		s.ValueType = "json"
	}
	return nil
}

// SettingRequest 设置请求结构
type SettingRequest struct {
	Category    string      `json:"category" binding:"required"`
	Key         string      `json:"key" binding:"required"`
	Value       interface{} `json:"value" binding:"required"`
	Description string      `json:"description"`
	Reason      string      `json:"reason"` // 修改原因
}

// SettingResponse 设置响应结构
type SettingResponse struct {
	ID          uuid.UUID   `json:"id"`
	Category    string      `json:"category"`
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	ValueType   string      `json:"value_type"`
	Description string      `json:"description"`
	IsActive    bool        `json:"is_active"`
	IsSystem    bool        `json:"is_system"`
	Version     int         `json:"version"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Creator     *User       `json:"creator,omitempty"`
	Updater     *User       `json:"updater,omitempty"`
}

// ToResponse 转换为响应结构
func (s *SystemSetting) ToResponse() *SettingResponse {
	response := &SettingResponse{
		ID:          s.ID,
		Category:    s.Category,
		Key:         s.Key,
		ValueType:   s.ValueType,
		Description: s.Description,
		IsActive:    s.IsActive,
		IsSystem:    s.IsSystem,
		Version:     s.Version,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}

	// 解析值
	switch s.ValueType {
	case "string":
		response.Value = s.Value
	case "int":
		if val, err := s.GetIntValue(); err == nil {
			response.Value = val
		}
	case "bool":
		if val, err := s.GetBoolValue(); err == nil {
			response.Value = val
		}
	case "json":
		var val interface{}
		if err := s.GetJSONValue(&val); err == nil {
			response.Value = val
		}
	}

	// 关联数据
	if s.Creator.ID != uuid.Nil {
		response.Creator = &s.Creator
	}
	if s.Updater.ID != uuid.Nil {
		response.Updater = &s.Updater
	}

	return response
}

// CategoryResponse 分类响应结构
type CategoryResponse struct {
	ID          uuid.UUID          `json:"id"`
	Name        string             `json:"name"`
	DisplayName string             `json:"display_name"`
	Description string             `json:"description"`
	Icon        string             `json:"icon"`
	SortOrder   int                `json:"sort_order"`
	IsActive    bool               `json:"is_active"`
	Settings    []*SettingResponse `json:"settings,omitempty"`
}

// ToResponse 转换为响应结构
func (c *SystemSettingCategory) ToResponse() *CategoryResponse {
	response := &CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		DisplayName: c.DisplayName,
		Description: c.Description,
		Icon:        c.Icon,
		SortOrder:   c.SortOrder,
		IsActive:    c.IsActive,
	}

	// 转换设置
	if len(c.Settings) > 0 {
		response.Settings = make([]*SettingResponse, len(c.Settings))
		for i, setting := range c.Settings {
			response.Settings[i] = setting.ToResponse()
		}
	}

	return response
}

// SettingsBackup 设置备份结构
type SettingsBackup struct {
	Version    string                      `json:"version"`
	Timestamp  time.Time                   `json:"timestamp"`
	Categories []*CategoryResponse         `json:"categories"`
	Settings   map[string]*SettingResponse `json:"settings"`
	Metadata   map[string]interface{}      `json:"metadata"`
}
