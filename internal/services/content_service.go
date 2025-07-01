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

type ContentService struct {
	geoipService *GeoIPService
	planService  *PlanService
}

func NewContentService(geoipService *GeoIPService) *ContentService {
	return &ContentService{
		geoipService: geoipService,
		planService:  NewPlanService(),
	}
}

type UploadRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Content     string     `json:"content" binding:"required"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// UpdateRequest 更新内容请求
type UpdateRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Content     string     `json:"content"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

func (s *ContentService) Upload(userID uuid.UUID, req *UploadRequest) (*models.Content, error) {
	// 检查用户使用限制
	limitStatus, err := s.planService.CheckUsageLimits(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check usage limits: %w", err)
	}

	if !limitStatus.CanUploadArticle {
		return nil, fmt.Errorf("monthly upload limit exceeded (%d/%d)", limitStatus.ArticlesUploaded, limitStatus.MonthlyUploadLimit)
	}

	// 根据用户等级计算过期时间
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		// 如果请求中指定了过期时间，使用指定的时间
		expiresAt = req.ExpiresAt
	} else {
		// 否则根据用户等级自动计算
		calculatedExpiration, err := s.planService.CalculateArticleExpiration(userID)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate expiration: %w", err)
		}
		expiresAt = calculatedExpiration
	}

	content := &models.Content{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
		ContentType: "text/html",
		ExpiresAt:   expiresAt,
		IsActive:    true,
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建内容
	if err := tx.Create(content).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create content: %w", err)
	}

	// 更新使用统计
	if err := s.updateUsageStatistics(tx, userID, 1, 0, 0); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update usage statistics: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return content, nil
}

// updateUsageStatistics 更新用户使用统计
func (s *ContentService) updateUsageStatistics(tx *gorm.DB, userID uuid.UUID, articlesUploaded int, storageUsedMB int64, apiCallsMade int) error {
	currentMonth := models.GetCurrentMonthYear()

	// PostgreSQL UPSERT
	err := tx.Exec(`
		INSERT INTO usage_statistics (user_id, month_year, articles_uploaded, storage_used_mb, api_calls_made, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
		ON CONFLICT (user_id, month_year)
		DO UPDATE SET
			articles_uploaded = usage_statistics.articles_uploaded + EXCLUDED.articles_uploaded,
			storage_used_mb = usage_statistics.storage_used_mb + EXCLUDED.storage_used_mb,
			api_calls_made = usage_statistics.api_calls_made + EXCLUDED.api_calls_made,
			updated_at = NOW()
	`, userID, currentMonth, articlesUploaded, storageUsedMB, apiCallsMade).Error

	return err
}

func (s *ContentService) GetByID(contentID uuid.UUID) (*models.Content, error) {
	var content models.Content
	if err := database.DB.Where("id = ? AND is_active = ?", contentID, true).First(&content).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("content not found")
		}
		return nil, err
	}
	return &content, nil
}

// GetByUserID 根据用户ID和内容ID获取内容
func (s *ContentService) GetByUserID(userID, contentID uuid.UUID) (*models.Content, error) {
	var content models.Content
	if err := database.DB.Where("id = ? AND user_id = ? AND is_active = ?", contentID, userID, true).First(&content).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("content not found")
		}
		return nil, err
	}
	return &content, nil
}

func (s *ContentService) List(userID uuid.UUID, page, limit int) ([]models.Content, int64, error) {
	var contents []models.Content
	var total int64

	offset := (page - 1) * limit

	if err := database.DB.Model(&models.Content{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := database.DB.Where("user_id = ? AND is_active = ?", userID, true).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&contents).Error; err != nil {
		return nil, 0, err
	}

	return contents, total, nil
}

func (s *ContentService) Update(userID, contentID uuid.UUID, req *UpdateRequest) (*models.Content, error) {
	var content models.Content
	if err := database.DB.Where("id = ? AND user_id = ? AND is_active = ?", contentID, userID, true).First(&content).Error; err != nil {
		return nil, errors.New("content not found")
	}

	// 更新字段
	if req.Title != "" {
		content.Title = req.Title
	}
	if req.Description != "" {
		content.Description = req.Description
	}
	if req.Content != "" {
		content.Content = req.Content
	}
	if req.ExpiresAt != nil {
		content.ExpiresAt = req.ExpiresAt
	}

	if err := database.DB.Save(&content).Error; err != nil {
		return nil, fmt.Errorf("failed to update content: %w", err)
	}

	return &content, nil
}

func (s *ContentService) Delete(userID, contentID uuid.UUID) error {
	result := database.DB.Model(&models.Content{}).
		Where("id = ? AND user_id = ?", contentID, userID).
		Update("is_active", false)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("content not found")
	}

	return nil
}

func (s *ContentService) ViewContent(contentID uuid.UUID, accessCode string, clientIP string) (*models.Content, error) {
	content, err := s.GetByID(contentID)
	if err != nil {
		return nil, err
	}

	if !content.CanAccess() {
		return nil, errors.New("access denied")
	}

	// 增加访问计数
	database.DB.Model(content).UpdateColumn("access_count", gorm.Expr("access_count + ?", 1))

	// 记录访问统计
	analytics := &models.ContentAnalytics{
		ContentID:  contentID,
		UserID:     content.UserID,
		IPAddress:  clientIP,
		AccessTime: time.Now(),
	}
	database.DB.Create(analytics)

	return content, nil
}

// ViewContentWithAnalytics 查看内容并记录详细的访问统计
func (s *ContentService) ViewContentWithAnalytics(contentID uuid.UUID, accessCode string, clientIP string, userAgent string, referer string) (*models.Content, error) {
	content, err := s.GetByID(contentID)
	if err != nil {
		return nil, err
	}

	if !content.CanAccess() {
		return nil, errors.New("access denied")
	}

	// 增加访问计数
	database.DB.Model(content).UpdateColumn("access_count", gorm.Expr("access_count + ?", 1))

	// 异步记录详细的访问统计
	go s.recordAnalyticsAsync(contentID, content.UserID, clientIP, userAgent, referer)

	return content, nil
}

// recordAnalyticsAsync 异步记录访问统计信息
func (s *ContentService) recordAnalyticsAsync(contentID, userID uuid.UUID, clientIP, userAgent, referer string) {
	// 记录详细的访问统计
	analytics := &models.ContentAnalytics{
		ContentID:  contentID,
		UserID:     userID,
		IPAddress:  clientIP,
		UserAgent:  userAgent,
		Referer:    referer,
		AccessTime: time.Now(),
	}

	// 使用 GeoIP 服务获取地理位置信息
	if s.geoipService != nil {
		if locationInfo, err := s.geoipService.GetLocationInfo(clientIP); err == nil {
			analytics.Country = locationInfo.Country
			analytics.Region = locationInfo.Country // 暂时使用国家作为地区
			analytics.City = locationInfo.City
			analytics.Latitude = locationInfo.Latitude
			analytics.Longitude = locationInfo.Longitude
		} else {
			// GeoIP 查询失败，使用默认值
			fmt.Printf("GeoIP lookup failed for %s: %v\n", clientIP, err)
			analytics.Country = "Unknown"
			analytics.Region = "Unknown"
			analytics.City = "Unknown"
		}
	} else {
		// GeoIP 服务不可用，使用简单的本地检测
		if clientIP == "127.0.0.1" || clientIP == "::1" {
			analytics.Country = "Local"
			analytics.Region = "Local"
			analytics.City = "Local"
		} else {
			analytics.Country = "Unknown"
			analytics.Region = "Unknown"
			analytics.City = "Unknown"
		}
	}

	if err := database.DB.Create(analytics).Error; err != nil {
		// 记录统计失败不应该影响内容访问
		fmt.Printf("Failed to record analytics: %v\n", err)
	}
}
