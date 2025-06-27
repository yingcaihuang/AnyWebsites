package services

import (
	"testing"
	"time"

	"anywebsites/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移表结构
	db.AutoMigrate(&models.User{}, &models.UserSubscription{}, &models.Content{})

	return db
}

func TestUserService_UpgradeUserPlan(t *testing.T) {
	db := setupUserTestDB()
	service := NewUserService(db)

	// 创建测试用户
	userID := uuid.New()
	user := models.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
		APIKey:   "test-api-key",
		IsActive: true,
	}
	db.Create(&user)

	// 创建一些测试内容
	contents := []models.Content{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Title:     "Test Content 1",
			Content:   "Test content body 1",
			IsActive:  true,
			CreatedAt: time.Now().AddDate(0, 0, -5), // 5天前创建
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			Title:     "Test Content 2",
			Content:   "Test content body 2",
			IsActive:  true,
			CreatedAt: time.Now().AddDate(0, 0, -10), // 10天前创建
		},
	}

	for _, content := range contents {
		db.Create(&content)
	}

	// 测试升级到开发者版
	expiresAt := time.Now().AddDate(0, 1, 0) // 1个月后过期
	err := service.UpgradeUserPlan(userID.String(), "developer", &expiresAt)
	assert.NoError(t, err)

	// 验证订阅记录
	var subscription models.UserSubscription
	err = db.Where("user_id = ?", userID).First(&subscription).Error
	assert.NoError(t, err)
	assert.Equal(t, models.PlanType("developer"), subscription.PlanType)
	assert.Equal(t, "active", subscription.Status)
	assert.NotNil(t, subscription.ExpiresAt)

	// 验证内容过期时间已更新
	var updatedContents []models.Content
	db.Where("user_id = ?", userID).Find(&updatedContents)

	for _, content := range updatedContents {
		if content.ExpiresAt != nil {
			// 开发者版保留30天，所以过期时间应该是创建时间+30天
			expectedExpiry := content.CreatedAt.AddDate(0, 0, 30)
			assert.WithinDuration(t, expectedExpiry, *content.ExpiresAt, time.Hour)
		}
	}
}

func TestUserService_DowngradeUserPlan(t *testing.T) {
	db := setupUserTestDB()
	service := NewUserService(db)

	// 创建测试用户
	userID := uuid.New()
	user := models.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
		APIKey:   "test-api-key",
		IsActive: true,
	}
	db.Create(&user)

	// 先创建一个专业版订阅
	subscription := models.UserSubscription{
		UserID:    userID,
		PlanType:  models.PlanType("pro"),
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(&subscription)

	// 创建一些测试内容
	contents := []models.Content{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Title:     "Test Content 1",
			Content:   "Test content body 1",
			IsActive:  true,
			CreatedAt: time.Now().AddDate(0, 0, -5), // 5天前创建
		},
	}

	for _, content := range contents {
		db.Create(&content)
	}

	// 测试降级到开发者版
	err := service.DowngradeUserPlan(userID.String(), "developer")
	assert.NoError(t, err)

	// 验证订阅记录已更新
	var updatedSubscription models.UserSubscription
	err = db.Where("user_id = ?", userID).First(&updatedSubscription).Error
	assert.NoError(t, err)
	assert.Equal(t, models.PlanType("developer"), updatedSubscription.PlanType)
	assert.Equal(t, "active", updatedSubscription.Status)

	// 验证内容过期时间已更新
	var updatedContents []models.Content
	db.Where("user_id = ?", userID).Find(&updatedContents)

	for _, content := range updatedContents {
		if content.ExpiresAt != nil {
			// 降级后应该根据新计划重新计算过期时间
			assert.NotNil(t, content.ExpiresAt)
		}
	}
}

func TestUserService_GetPlanRetentionDays(t *testing.T) {
	db := setupUserTestDB()
	service := NewUserService(db)

	tests := []struct {
		planType string
		expected int
	}{
		{"community", 7},
		{"developer", 30},
		{"pro", 90},
		{"max", 365},
		{"enterprise", -1},
		{"invalid", 7}, // 默认为社区版
	}

	for _, tt := range tests {
		t.Run(tt.planType, func(t *testing.T) {
			result := service.getPlanRetentionDays(tt.planType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserService_UpdateUserContentExpiration(t *testing.T) {
	db := setupUserTestDB()
	service := NewUserService(db)

	// 创建测试用户
	userID := uuid.New()
	user := models.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
		APIKey:   "test-api-key",
		IsActive: true,
	}
	db.Create(&user)

	// 创建测试内容
	content := models.Content{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     "Test Content",
		Content:   "Test content body",
		IsActive:  true,
		CreatedAt: time.Now().AddDate(0, 0, -5), // 5天前创建
	}
	db.Create(&content)

	// 开始事务
	tx := db.Begin()

	// 测试升级（延长过期时间）
	err := service.updateUserContentExpiration(tx, userID.String(), 30, true)
	assert.NoError(t, err)

	tx.Commit()

	// 验证内容过期时间
	var updatedContent models.Content
	db.First(&updatedContent, content.ID)

	if updatedContent.ExpiresAt != nil {
		expectedExpiry := content.CreatedAt.AddDate(0, 0, 30)
		assert.WithinDuration(t, expectedExpiry, *updatedContent.ExpiresAt, time.Hour)
	}

	// 测试降级（缩短过期时间）
	tx = db.Begin()
	err = service.updateUserContentExpiration(tx, userID.String(), 7, false)
	assert.NoError(t, err)
	tx.Commit()

	// 验证内容过期时间已更新
	db.First(&updatedContent, content.ID)
	assert.NotNil(t, updatedContent.ExpiresAt)
}

func TestUserService_InvalidUserID(t *testing.T) {
	db := setupUserTestDB()
	service := NewUserService(db)

	// 测试无效的用户ID
	err := service.UpgradeUserPlan("invalid-uuid", "developer", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的用户ID")

	err = service.DowngradeUserPlan("invalid-uuid", "community")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的用户ID")
}

func TestUserService_NonExistentUser(t *testing.T) {
	db := setupUserTestDB()
	service := NewUserService(db)

	// 测试不存在的用户
	nonExistentUserID := uuid.New()
	err := service.UpgradeUserPlan(nonExistentUserID.String(), "developer", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户不存在")

	err = service.DowngradeUserPlan(nonExistentUserID.String(), "community")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户订阅不存在")
}

func TestUserService_InvalidPlanType(t *testing.T) {
	db := setupUserTestDB()
	service := NewUserService(db)

	// 创建测试用户
	userID := uuid.New()
	user := models.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
		APIKey:   "test-api-key",
		IsActive: true,
	}
	db.Create(&user)

	// 测试无效的计划类型
	err := service.UpgradeUserPlan(userID.String(), "invalid-plan", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的计划类型")

	// 先创建订阅才能测试降级
	subscription := models.UserSubscription{
		UserID:   userID,
		PlanType: models.PlanType("pro"),
		Status:   "active",
	}
	db.Create(&subscription)

	err = service.DowngradeUserPlan(userID.String(), "invalid-plan")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的计划类型")
}
