package services

import (
	"fmt"
	"sync"
	"time"

	"anywebsites/internal/database"
	"anywebsites/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SettingsService 系统设置服务
type SettingsService struct {
	cache       map[string]*models.SystemSetting
	cacheMutex  sync.RWMutex
	cacheExpiry time.Duration
	lastUpdate  time.Time
}

// NewSettingsService 创建设置服务实例
func NewSettingsService() *SettingsService {
	service := &SettingsService{
		cache:       make(map[string]*models.SystemSetting),
		cacheExpiry: 5 * time.Minute, // 缓存5分钟
	}

	// 初始加载设置
	service.RefreshCache()

	return service
}

// GetSetting 获取设置值
func (s *SettingsService) GetSetting(category, key string) (*models.SystemSetting, error) {
	cacheKey := fmt.Sprintf("%s.%s", category, key)

	// 先从缓存获取
	s.cacheMutex.RLock()
	if setting, exists := s.cache[cacheKey]; exists {
		// 检查缓存是否过期
		if time.Since(s.lastUpdate) < s.cacheExpiry {
			s.cacheMutex.RUnlock()
			return setting, nil
		}
	}
	s.cacheMutex.RUnlock()

	// 从数据库获取
	var setting models.SystemSetting
	err := database.DB.Where("category = ? AND key = ? AND is_active = ?", category, key, true).
		First(&setting).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("setting not found: %s.%s", category, key)
		}
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}

	// 更新缓存
	s.cacheMutex.Lock()
	s.cache[cacheKey] = &setting
	s.cacheMutex.Unlock()

	return &setting, nil
}

// GetStringValue 获取字符串设置值
func (s *SettingsService) GetStringValue(category, key, defaultValue string) string {
	setting, err := s.GetSetting(category, key)
	if err != nil {
		return defaultValue
	}
	return setting.GetStringValue()
}

// GetIntValue 获取整数设置值
func (s *SettingsService) GetIntValue(category, key string, defaultValue int) int {
	setting, err := s.GetSetting(category, key)
	if err != nil {
		return defaultValue
	}

	if value, err := setting.GetIntValue(); err == nil {
		return value
	}
	return defaultValue
}

// GetBoolValue 获取布尔设置值
func (s *SettingsService) GetBoolValue(category, key string, defaultValue bool) bool {
	setting, err := s.GetSetting(category, key)
	if err != nil {
		return defaultValue
	}

	if value, err := setting.GetBoolValue(); err == nil {
		return value
	}
	return defaultValue
}

// GetJSONValue 获取JSON设置值
func (s *SettingsService) GetJSONValue(category, key string, target interface{}) error {
	setting, err := s.GetSetting(category, key)
	if err != nil {
		return err
	}

	return setting.GetJSONValue(target)
}

// SetSetting 设置值
func (s *SettingsService) SetSetting(category, key string, value interface{}, description string, userID uuid.UUID, reason string) error {
	// 验证输入
	if err := s.validateSetting(category, key, value); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查找现有设置
	var existingSetting models.SystemSetting
	err := tx.Where("category = ? AND key = ?", category, key).First(&existingSetting).Error

	var oldValue string
	changeType := "create"

	if err == nil {
		// 更新现有设置
		oldValue = existingSetting.Value
		changeType = "update"

		// 设置新值
		if err := existingSetting.SetValue(value); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to set value: %w", err)
		}

		existingSetting.Description = description
		existingSetting.UpdatedBy = userID
		existingSetting.Version++

		if err := tx.Save(&existingSetting).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update setting: %w", err)
		}

	} else if err == gorm.ErrRecordNotFound {
		// 创建新设置
		newSetting := models.SystemSetting{
			Category:    category,
			Key:         key,
			Description: description,
			IsActive:    true,
			IsSystem:    false,
			Version:     1,
			CreatedBy:   userID,
			UpdatedBy:   userID,
		}

		if err := newSetting.SetValue(value); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to set value: %w", err)
		}

		if err := tx.Create(&newSetting).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create setting: %w", err)
		}

		existingSetting = newSetting
	} else {
		tx.Rollback()
		return fmt.Errorf("failed to query setting: %w", err)
	}

	// 记录历史
	history := models.SystemSettingHistory{
		SettingID:  existingSetting.ID,
		OldValue:   oldValue,
		NewValue:   existingSetting.Value,
		ChangeType: changeType,
		Reason:     reason,
		CreatedBy:  userID,
	}

	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create history: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 清除缓存
	s.invalidateCache(category, key)

	return nil
}

// DeleteSetting 删除设置
func (s *SettingsService) DeleteSetting(category, key string, userID uuid.UUID, reason string) error {
	// 查找设置
	var setting models.SystemSetting
	err := database.DB.Where("category = ? AND key = ?", category, key).First(&setting).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("setting not found: %s.%s", category, key)
		}
		return fmt.Errorf("failed to find setting: %w", err)
	}

	// 检查是否为系统设置
	if setting.IsSystem {
		return fmt.Errorf("cannot delete system setting: %s.%s", category, key)
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 记录历史
	history := models.SystemSettingHistory{
		SettingID:  setting.ID,
		OldValue:   setting.Value,
		NewValue:   "",
		ChangeType: "delete",
		Reason:     reason,
		CreatedBy:  userID,
	}

	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create history: %w", err)
	}

	// 软删除设置
	setting.IsActive = false
	setting.UpdatedBy = userID

	if err := tx.Save(&setting).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete setting: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 清除缓存
	s.invalidateCache(category, key)

	return nil
}

// GetAllSettings 获取所有设置
func (s *SettingsService) GetAllSettings() ([]*models.SettingResponse, error) {
	var settings []models.SystemSetting
	err := database.DB.Where("is_active = ?", true).
		Preload("Creator").
		Preload("Updater").
		Order("category, key").
		Find(&settings).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	responses := make([]*models.SettingResponse, len(settings))
	for i, setting := range settings {
		responses[i] = setting.ToResponse()
	}

	return responses, nil
}

// GetSettingsByCategory 按分类获取设置
func (s *SettingsService) GetSettingsByCategory(category string) ([]*models.SettingResponse, error) {
	var settings []models.SystemSetting
	err := database.DB.Where("category = ? AND is_active = ?", category, true).
		Preload("Creator").
		Preload("Updater").
		Order("key").
		Find(&settings).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	responses := make([]*models.SettingResponse, len(settings))
	for i, setting := range settings {
		responses[i] = setting.ToResponse()
	}

	return responses, nil
}

// GetCategories 获取所有分类
func (s *SettingsService) GetCategories() ([]*models.CategoryResponse, error) {
	var categories []models.SystemSettingCategory
	err := database.DB.Where("is_active = ?", true).
		Preload("Settings", "is_active = ?", true).
		Preload("Settings.Creator").
		Preload("Settings.Updater").
		Order("sort_order, display_name").
		Find(&categories).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	responses := make([]*models.CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = category.ToResponse()
	}

	return responses, nil
}

// GetSettingHistory 获取设置历史
func (s *SettingsService) GetSettingHistory(category, key string, limit int) ([]models.SystemSettingHistory, error) {
	// 先找到设置
	var setting models.SystemSetting
	err := database.DB.Where("category = ? AND key = ?", category, key).First(&setting).Error
	if err != nil {
		return nil, fmt.Errorf("setting not found: %w", err)
	}

	// 获取历史记录
	var histories []models.SystemSettingHistory
	query := database.DB.Where("setting_id = ?", setting.ID).
		Preload("Creator").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err = query.Find(&histories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	return histories, nil
}

// RefreshCache 刷新缓存
func (s *SettingsService) RefreshCache() error {
	var settings []models.SystemSetting
	err := database.DB.Where("is_active = ?", true).Find(&settings).Error
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// 清空缓存
	s.cache = make(map[string]*models.SystemSetting)

	// 重新加载
	for _, setting := range settings {
		cacheKey := fmt.Sprintf("%s.%s", setting.Category, setting.Key)
		settingCopy := setting // 避免指针问题
		s.cache[cacheKey] = &settingCopy
	}

	s.lastUpdate = time.Now()
	return nil
}

// invalidateCache 清除特定设置的缓存
func (s *SettingsService) invalidateCache(category, key string) {
	cacheKey := fmt.Sprintf("%s.%s", category, key)

	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	delete(s.cache, cacheKey)
}

// validateSetting 验证设置
func (s *SettingsService) validateSetting(category, key string, value interface{}) error {
	// 基本验证
	if category == "" {
		return fmt.Errorf("category cannot be empty")
	}
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// 特定验证规则
	switch category {
	case "server":
		return s.validateServerSetting(key, value)
	case "database":
		return s.validateDatabaseSetting(key, value)
	case "upload":
		return s.validateUploadSetting(key, value)
	case "security":
		return s.validateSecuritySetting(key, value)
	}

	return nil
}

// validateServerSetting 验证服务器设置
func (s *SettingsService) validateServerSetting(key string, value interface{}) error {
	switch key {
	case "port", "metrics_port", "health_check_port", "pprof_port":
		// 处理JSON解析时的float64类型
		var port int
		switch v := value.(type) {
		case int:
			port = v
		case float64:
			port = int(v)
		default:
			return fmt.Errorf("%s must be an integer", key)
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("%s must be between 1 and 65535", key)
		}
	case "host":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("host must be a string")
		}
	}
	return nil
}

// validateDatabaseSetting 验证数据库设置
func (s *SettingsService) validateDatabaseSetting(key string, value interface{}) error {
	switch key {
	case "port":
		// 处理JSON解析时的float64类型
		var port int
		switch v := value.(type) {
		case int:
			port = v
		case float64:
			port = int(v)
		default:
			return fmt.Errorf("database port must be an integer")
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("database port must be between 1 and 65535")
		}
	case "host", "user", "password", "name":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%s must be a string", key)
		}
	}
	return nil
}

// validateUploadSetting 验证上传设置
func (s *SettingsService) validateUploadSetting(key string, value interface{}) error {
	switch key {
	case "max_file_size":
		if size, ok := value.(int); ok {
			if size < 1024 || size > 1024*1024*1024 { // 1KB to 1GB
				return fmt.Errorf("max file size must be between 1KB and 1GB")
			}
		} else {
			return fmt.Errorf("max file size must be an integer")
		}
	case "path":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("upload path must be a string")
		}
	}
	return nil
}

// validateSecuritySetting 验证安全设置
func (s *SettingsService) validateSecuritySetting(key string, value interface{}) error {
	switch key {
	case "jwt_secret":
		if secret, ok := value.(string); ok {
			if len(secret) < 32 {
				return fmt.Errorf("JWT secret must be at least 32 characters")
			}
		} else {
			return fmt.Errorf("JWT secret must be a string")
		}
	case "rate_limit_requests":
		if requests, ok := value.(int); ok {
			if requests < 1 || requests > 10000 {
				return fmt.Errorf("rate limit requests must be between 1 and 10000")
			}
		} else {
			return fmt.Errorf("rate limit requests must be an integer")
		}
	}
	return nil
}

// ExportSettings 导出设置
func (s *SettingsService) ExportSettings() (*models.SettingsBackup, error) {
	categories, err := s.GetCategories()
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	settings, err := s.GetAllSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	// 构建设置映射
	settingsMap := make(map[string]*models.SettingResponse)
	for _, setting := range settings {
		key := fmt.Sprintf("%s.%s", setting.Category, setting.Key)
		settingsMap[key] = setting
	}

	backup := &models.SettingsBackup{
		Version:    "1.0",
		Timestamp:  time.Now(),
		Categories: categories,
		Settings:   settingsMap,
		Metadata: map[string]interface{}{
			"total_settings":   len(settings),
			"total_categories": len(categories),
			"exported_by":      "system",
		},
	}

	return backup, nil
}

// ImportSettings 导入设置
func (s *SettingsService) ImportSettings(backup *models.SettingsBackup, userID uuid.UUID, overwrite bool) error {
	if backup == nil {
		return fmt.Errorf("backup data is nil")
	}

	// 验证备份版本
	if backup.Version != "1.0" {
		return fmt.Errorf("unsupported backup version: %s", backup.Version)
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 导入设置
	for key, settingData := range backup.Settings {
		category := settingData.Category
		settingKey := settingData.Key

		// 检查是否已存在
		var existingSetting models.SystemSetting
		err := tx.Where("category = ? AND key = ?", category, settingKey).First(&existingSetting).Error

		if err == gorm.ErrRecordNotFound {
			// 创建新设置
			newSetting := models.SystemSetting{
				Category:    category,
				Key:         settingKey,
				Description: settingData.Description,
				IsActive:    true,
				IsSystem:    false,
				Version:     1,
				CreatedBy:   userID,
				UpdatedBy:   userID,
			}

			if err := newSetting.SetValue(settingData.Value); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to set value for %s: %w", key, err)
			}

			if err := tx.Create(&newSetting).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create setting %s: %w", key, err)
			}

		} else if err == nil && overwrite {
			// 更新现有设置
			if err := existingSetting.SetValue(settingData.Value); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to set value for %s: %w", key, err)
			}

			existingSetting.Description = settingData.Description
			existingSetting.UpdatedBy = userID
			existingSetting.Version++

			if err := tx.Save(&existingSetting).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update setting %s: %w", key, err)
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit import: %w", err)
	}

	// 刷新缓存
	s.RefreshCache()

	return nil
}
