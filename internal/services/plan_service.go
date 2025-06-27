package services

import (
	"errors"
	"fmt"
	"time"

	"anywebsites/internal/database"
	"anywebsites/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlanService struct{}

func NewPlanService() *PlanService {
	return &PlanService{}
}

// GetUserPlan 获取用户当前计划
func (s *PlanService) GetUserPlan(userID uuid.UUID) (*models.UserSubscription, error) {
	var subscription models.UserSubscription
	err := database.DB.Where("user_id = ? AND status = ?", userID, models.StatusActive).
		First(&subscription).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果没有找到订阅，创建默认的社区版订阅
			return s.CreateDefaultSubscription(userID)
		}
		return nil, fmt.Errorf("failed to get user plan: %w", err)
	}

	// 检查是否过期
	if subscription.IsExpired() {
		// 过期则降级为社区版
		return s.DowngradeToFree(userID)
	}

	return &subscription, nil
}

// CreateDefaultSubscription 创建默认订阅（社区版）
func (s *PlanService) CreateDefaultSubscription(userID uuid.UUID) (*models.UserSubscription, error) {
	subscription := &models.UserSubscription{
		UserID:    userID,
		PlanType:  models.PlanCommunity,
		Status:    models.StatusActive,
		StartedAt: time.Now(),
		// 社区版不设置过期时间
	}

	if err := database.DB.Create(subscription).Error; err != nil {
		return nil, fmt.Errorf("failed to create default subscription: %w", err)
	}

	// 预加载计划配置
	if err := database.DB.Preload("PlanConfig").First(subscription, subscription.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load plan config: %w", err)
	}

	return subscription, nil
}

// GetPlanConfig 获取计划配置
func (s *PlanService) GetPlanConfig(planType models.PlanType) (*models.PlanConfig, error) {
	var config models.PlanConfig
	err := database.DB.Where("type = ? AND is_active = ?", planType, true).First(&config).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get plan config: %w", err)
	}
	return &config, nil
}

// CalculateArticleExpiration 根据用户计划计算文章过期时间
func (s *PlanService) CalculateArticleExpiration(userID uuid.UUID) (*time.Time, error) {
	subscription, err := s.GetUserPlan(userID)
	if err != nil {
		return nil, err
	}

	// 通过计划类型获取配置
	config, err := s.GetPlanConfig(subscription.PlanType)
	if err != nil {
		return nil, err
	}

	// 如果是无限制计划（企业版）
	if config.ArticleRetentionDays == -1 {
		return nil, nil // 返回 nil 表示永不过期
	}

	// 计算过期时间
	expirationTime := time.Now().AddDate(0, 0, config.ArticleRetentionDays)
	return &expirationTime, nil
}

// UpdateArticleExpirations 更新用户所有文章的过期时间（用于计划变更时）
func (s *PlanService) UpdateArticleExpirations(userID uuid.UUID, newPlanType models.PlanType) error {
	// 获取新计划配置
	newConfig, err := s.GetPlanConfig(newPlanType)
	if err != nil {
		return err
	}

	// 获取用户所有活跃文章
	var contents []models.Content
	if err := database.DB.Where("user_id = ? AND is_active = ?", userID, true).Find(&contents).Error; err != nil {
		return fmt.Errorf("failed to get user contents: %w", err)
	}

	// 批量更新过期时间
	for _, content := range contents {
		var newExpiration *time.Time

		if newConfig.ArticleRetentionDays == -1 {
			// 无限制计划，设置为永不过期
			newExpiration = nil
		} else {
			// 从文章创建时间开始计算新的过期时间
			expTime := content.CreatedAt.AddDate(0, 0, newConfig.ArticleRetentionDays)
			newExpiration = &expTime
		}

		// 更新文章过期时间
		if err := database.DB.Model(&content).Update("expires_at", newExpiration).Error; err != nil {
			return fmt.Errorf("failed to update content expiration: %w", err)
		}
	}

	return nil
}

// UpgradePlan 升级用户计划
func (s *PlanService) UpgradePlan(userID uuid.UUID, newPlanType models.PlanType, expiresAt *time.Time) error {
	// 获取当前订阅
	currentSubscription, err := s.GetUserPlan(userID)
	if err != nil {
		return err
	}

	// 记录升级历史
	history := &models.PlanUpgradeHistory{
		UserID:       userID,
		FromPlan:     currentSubscription.PlanType,
		ToPlan:       newPlanType,
		ChangeReason: "upgrade",
		EffectiveAt:  time.Now(),
		Status:       models.StatusActive,
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新订阅
	if err := tx.Model(currentSubscription).Updates(map[string]interface{}{
		"plan_type":  newPlanType,
		"expires_at": expiresAt,
		"status":     models.StatusActive,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// 记录历史
	if err := tx.Create(history).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create upgrade history: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 更新文章过期时间
	if err := s.UpdateArticleExpirations(userID, newPlanType); err != nil {
		// 这里只记录错误，不回滚订阅更新
		fmt.Printf("Warning: failed to update article expirations: %v\n", err)
	}

	return nil
}

// DowngradeToFree 降级到免费版
func (s *PlanService) DowngradeToFree(userID uuid.UUID) (*models.UserSubscription, error) {
	// 获取当前订阅
	var currentSubscription models.UserSubscription
	if err := database.DB.Where("user_id = ?", userID).First(&currentSubscription).Error; err != nil {
		// 如果没有订阅记录，创建新的免费版订阅
		return s.CreateDefaultSubscription(userID)
	}

	// 记录降级历史
	history := &models.PlanUpgradeHistory{
		UserID:       userID,
		FromPlan:     currentSubscription.PlanType,
		ToPlan:       models.PlanCommunity,
		ChangeReason: "downgrade_expired",
		EffectiveAt:  time.Now(),
		Status:       models.StatusActive,
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新订阅为免费版
	if err := tx.Model(&currentSubscription).Updates(map[string]interface{}{
		"plan_type":  models.PlanCommunity,
		"expires_at": nil, // 免费版不过期
		"status":     models.StatusActive,
	}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to downgrade subscription: %w", err)
	}

	// 记录历史
	if err := tx.Create(history).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create downgrade history: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 更新文章过期时间
	if err := s.UpdateArticleExpirations(userID, models.PlanCommunity); err != nil {
		fmt.Printf("Warning: failed to update article expirations: %v\n", err)
	}

	// 重新加载订阅信息
	if err := database.DB.Preload("PlanConfig").First(&currentSubscription, currentSubscription.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload subscription: %w", err)
	}

	return &currentSubscription, nil
}

// CheckUsageLimits 检查用户使用限制
func (s *PlanService) CheckUsageLimits(userID uuid.UUID) (*UsageLimitStatus, error) {
	subscription, err := s.GetUserPlan(userID)
	if err != nil {
		return nil, err
	}

	// 通过计划类型获取配置
	config, err := s.GetPlanConfig(subscription.PlanType)
	if err != nil {
		return nil, err
	}

	currentMonth := models.GetCurrentMonthYear()

	// 获取当月使用统计
	var usage models.UsageStatistics
	err = database.DB.Where("user_id = ? AND month_year = ?", userID, currentMonth).First(&usage).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to get usage statistics: %w", err)
	}

	status := &UsageLimitStatus{
		PlanType:            config.Type,
		MonthlyUploadLimit:  config.MonthlyUploadLimit,
		StorageLimitMB:      config.StorageLimitMB,
		APIRateLimitPerHour: config.APIRateLimitPerHour,
		ArticlesUploaded:    usage.ArticlesUploaded,
		StorageUsedMB:       usage.StorageUsedMB,
		APICallsMade:        usage.APICallsMade,
		CanUploadArticle:    true,
		CanMakeAPICall:      true,
		HasStorageSpace:     true,
	}

	// 检查限制（企业版无限制）
	if config.Type != models.PlanEnterprise {
		if config.MonthlyUploadLimit > 0 && usage.ArticlesUploaded >= config.MonthlyUploadLimit {
			status.CanUploadArticle = false
		}
		if config.StorageLimitMB > 0 && usage.StorageUsedMB >= config.StorageLimitMB {
			status.HasStorageSpace = false
		}
		// API 频率限制需要在中间件中实现
	}

	return status, nil
}

// GetPlanHistory 获取用户计划变更历史
func (s *PlanService) GetPlanHistory(userID uuid.UUID, histories *[]models.PlanUpgradeHistory) error {
	err := database.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(50). // 限制返回最近50条记录
		Find(histories).Error

	if err != nil {
		return fmt.Errorf("failed to get plan history: %w", err)
	}

	return nil
}

// UsageLimitStatus 使用限制状态
type UsageLimitStatus struct {
	PlanType            models.PlanType `json:"plan_type"`
	MonthlyUploadLimit  int             `json:"monthly_upload_limit"`
	StorageLimitMB      int64           `json:"storage_limit_mb"`
	APIRateLimitPerHour int             `json:"api_rate_limit_per_hour"`
	ArticlesUploaded    int             `json:"articles_uploaded"`
	StorageUsedMB       int64           `json:"storage_used_mb"`
	APICallsMade        int             `json:"api_calls_made"`
	CanUploadArticle    bool            `json:"can_upload_article"`
	CanMakeAPICall      bool            `json:"can_make_api_call"`
	HasStorageSpace     bool            `json:"has_storage_space"`
}
