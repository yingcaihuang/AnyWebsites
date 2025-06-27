package services

import (
	"fmt"
	"log"
	"time"

	"anywebsites/internal/database"
	"anywebsites/internal/models"
)

type CleanupService struct {
	stopChan chan bool
}

func NewCleanupService() *CleanupService {
	return &CleanupService{
		stopChan: make(chan bool),
	}
}

// Start 启动清理服务
func (s *CleanupService) Start() {
	log.Println("🧹 Starting cleanup service...")

	// 立即执行一次清理
	s.runCleanup()

	// 设置定时器，每小时执行一次
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runCleanup()
		case <-s.stopChan:
			log.Println("🛑 Cleanup service stopped")
			return
		}
	}
}

// Stop 停止清理服务
func (s *CleanupService) Stop() {
	close(s.stopChan)
}

// runCleanup 执行清理任务
func (s *CleanupService) runCleanup() {
	log.Println("🧹 Running cleanup tasks...")

	// 1. 软删除过期文章
	if err := s.softDeleteExpiredContent(); err != nil {
		log.Printf("❌ Error soft deleting expired content: %v", err)
	}

	// 2. 硬删除软删除超过30天的文章
	if err := s.hardDeleteOldContent(); err != nil {
		log.Printf("❌ Error hard deleting old content: %v", err)
	}

	// 3. 清理过期的用户订阅
	if err := s.cleanupExpiredSubscriptions(); err != nil {
		log.Printf("❌ Error cleaning up expired subscriptions: %v", err)
	}

	// 4. 清理旧的使用统计数据（保留12个月）
	if err := s.cleanupOldUsageStatistics(); err != nil {
		log.Printf("❌ Error cleaning up old usage statistics: %v", err)
	}

	log.Println("✅ Cleanup tasks completed")
}

// softDeleteExpiredContent 软删除过期文章
func (s *CleanupService) softDeleteExpiredContent() error {
	now := time.Now()

	// 查找所有过期且仍然活跃的文章
	var expiredContents []models.Content
	err := database.DB.Where("expires_at IS NOT NULL AND expires_at < ? AND is_active = ?", now, true).Find(&expiredContents).Error
	if err != nil {
		return fmt.Errorf("failed to find expired content: %w", err)
	}

	if len(expiredContents) == 0 {
		log.Println("📄 No expired content found")
		return nil
	}

	// 批量软删除
	result := database.DB.Model(&models.Content{}).
		Where("expires_at IS NOT NULL AND expires_at < ? AND is_active = ?", now, true).
		Updates(map[string]interface{}{
			"is_active":  false,
			"deleted_at": now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to soft delete expired content: %w", result.Error)
	}

	log.Printf("🗑️ Soft deleted %d expired articles", result.RowsAffected)

	// 记录清理统计
	s.logCleanupStats("soft_delete", int(result.RowsAffected))

	return nil
}

// hardDeleteOldContent 硬删除软删除超过30天的文章
func (s *CleanupService) hardDeleteOldContent() error {
	// 30天前的时间
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	// 查找软删除超过30天的文章
	var oldContents []models.Content
	err := database.DB.Unscoped().Where("is_active = ? AND deleted_at IS NOT NULL AND deleted_at < ?", false, thirtyDaysAgo).Find(&oldContents).Error
	if err != nil {
		return fmt.Errorf("failed to find old content: %w", err)
	}

	if len(oldContents) == 0 {
		log.Println("📄 No old content found for hard deletion")
		return nil
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除相关的分析数据
	var contentIDs []string
	for _, content := range oldContents {
		contentIDs = append(contentIDs, content.ID.String())
	}

	// 删除内容分析数据
	if err := tx.Where("content_id IN ?", contentIDs).Delete(&models.ContentAnalytics{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete content analytics: %w", err)
	}

	// 硬删除内容
	result := tx.Unscoped().Where("is_active = ? AND deleted_at IS NOT NULL AND deleted_at < ?", false, thirtyDaysAgo).Delete(&models.Content{})
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to hard delete old content: %w", result.Error)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit hard delete transaction: %w", err)
	}

	log.Printf("🗑️ Hard deleted %d old articles and their analytics", result.RowsAffected)

	// 记录清理统计
	s.logCleanupStats("hard_delete", int(result.RowsAffected))

	return nil
}

// cleanupExpiredSubscriptions 清理过期订阅
func (s *CleanupService) cleanupExpiredSubscriptions() error {
	now := time.Now()

	// 查找过期的订阅
	var expiredSubscriptions []models.UserSubscription
	err := database.DB.Where("expires_at IS NOT NULL AND expires_at < ? AND status = ?", now, models.StatusActive).Find(&expiredSubscriptions).Error
	if err != nil {
		return fmt.Errorf("failed to find expired subscriptions: %w", err)
	}

	if len(expiredSubscriptions) == 0 {
		log.Println("📋 No expired subscriptions found")
		return nil
	}

	planService := NewPlanService()

	// 处理每个过期订阅
	for _, subscription := range expiredSubscriptions {
		// 降级到免费版
		_, err := planService.DowngradeToFree(subscription.UserID)
		if err != nil {
			log.Printf("❌ Failed to downgrade user %s: %v", subscription.UserID, err)
			continue
		}

		log.Printf("⬇️ Downgraded user %s from %s to community plan", subscription.UserID, subscription.PlanType)
	}

	log.Printf("📋 Processed %d expired subscriptions", len(expiredSubscriptions))

	return nil
}

// cleanupOldUsageStatistics 清理旧的使用统计数据
func (s *CleanupService) cleanupOldUsageStatistics() error {
	// 保留12个月的数据
	twelveMonthsAgo := time.Now().AddDate(0, -12, 0)
	cutoffMonth := twelveMonthsAgo.Format("2006-01")

	result := database.DB.Where("month_year < ?", cutoffMonth).Delete(&models.UsageStatistics{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old usage statistics: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		log.Printf("📊 Cleaned up %d old usage statistics records", result.RowsAffected)
	}

	return nil
}

// logCleanupStats 记录清理统计
func (s *CleanupService) logCleanupStats(operation string, count int) {
	// 这里可以记录到数据库或发送到监控系统
	log.Printf("📈 Cleanup stats - Operation: %s, Count: %d, Time: %s", operation, count, time.Now().Format(time.RFC3339))
}

// GetCleanupStats 获取清理统计信息
func (s *CleanupService) GetCleanupStats() (*CleanupStats, error) {
	stats := &CleanupStats{}

	// 统计过期但未删除的文章
	err := database.DB.Model(&models.Content{}).
		Where("expires_at IS NOT NULL AND expires_at < ? AND is_active = ?", time.Now(), true).
		Count(&stats.ExpiredActiveContent).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count expired active content: %w", err)
	}

	// 统计软删除的文章
	err = database.DB.Model(&models.Content{}).
		Where("is_active = ? AND deleted_at IS NOT NULL", false).
		Count(&stats.SoftDeletedContent).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count soft deleted content: %w", err)
	}

	// 统计过期的订阅
	err = database.DB.Model(&models.UserSubscription{}).
		Where("expires_at IS NOT NULL AND expires_at < ? AND status = ?", time.Now(), models.StatusActive).
		Count(&stats.ExpiredSubscriptions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count expired subscriptions: %w", err)
	}

	// 统计总的活跃文章数
	err = database.DB.Model(&models.Content{}).
		Where("is_active = ?", true).
		Count(&stats.TotalActiveContent).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total active content: %w", err)
	}

	return stats, nil
}

// CleanupStats 清理统计信息
type CleanupStats struct {
	ExpiredActiveContent int64     `json:"expired_active_content"`
	SoftDeletedContent   int64     `json:"soft_deleted_content"`
	ExpiredSubscriptions int64     `json:"expired_subscriptions"`
	TotalActiveContent   int64     `json:"total_active_content"`
	LastCleanupTime      time.Time `json:"last_cleanup_time"`
}

// ManualCleanup 手动执行清理任务
func (s *CleanupService) ManualCleanup() error {
	log.Println("🔧 Manual cleanup triggered")
	s.runCleanup()
	return nil
}
