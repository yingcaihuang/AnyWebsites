package services

import (
	"testing"

	"anywebsites/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移表结构
	db.AutoMigrate(&models.User{}, &models.UserSubscription{}, &models.PlanConfig{})

	return db
}

func TestPlanService_GetAllPlans(t *testing.T) {
	service := NewPlanService()

	plans := service.GetAllPlans()

	assert.NotEmpty(t, plans)
	assert.Len(t, plans, 5) // community, developer, pro, max, enterprise

	// 验证计划类型
	planTypes := make(map[string]bool)
	for _, plan := range plans {
		planTypes[string(plan.Type)] = true
	}

	assert.True(t, planTypes["community"])
	assert.True(t, planTypes["developer"])
	assert.True(t, planTypes["pro"])
	assert.True(t, planTypes["max"])
	assert.True(t, planTypes["enterprise"])
}

func TestPlanService_GetPlanByType(t *testing.T) {
	service := NewPlanService()

	tests := []struct {
		planType string
		wantErr  bool
	}{
		{"community", false},
		{"developer", false},
		{"pro", false},
		{"max", false},
		{"enterprise", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.planType, func(t *testing.T) {
			plan, err := service.GetPlanByType(tt.planType)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, plan)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
				assert.Equal(t, tt.planType, string(plan.Type))
			}
		})
	}
}

func TestPlanService_UpgradePlan(t *testing.T) {
	db := setupTestDB()
	service := NewPlanService()

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

	// 测试获取计划信息
	plan, err := service.GetPlanByType("developer")
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, "developer", string(plan.Type))
	assert.Equal(t, 30, plan.ArticleRetentionDays)

	// 测试获取专业版计划
	plan, err = service.GetPlanByType("pro")
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, "pro", string(plan.Type))
	assert.Equal(t, 90, plan.ArticleRetentionDays)
}

func TestPlanService_ValidatePlanTypes(t *testing.T) {
	service := NewPlanService()

	// 测试有效的计划类型
	validTypes := []string{"community", "developer", "pro", "max", "enterprise"}
	for _, planType := range validTypes {
		plan, err := service.GetPlanByType(planType)
		assert.NoError(t, err)
		assert.NotNil(t, plan)
		assert.Equal(t, planType, string(plan.Type))
	}

	// 测试无效的计划类型
	_, err := service.GetPlanByType("invalid")
	assert.Error(t, err)
}

func TestPlanService_PlanFeatures(t *testing.T) {
	service := NewPlanService()

	// 测试社区版特性
	plan, err := service.GetPlanByType("community")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, plan.Price)
	assert.Equal(t, 7, plan.ArticleRetentionDays)
	assert.Equal(t, 50, plan.MonthlyUploadLimit)

	// 测试开发者版特性
	plan, err = service.GetPlanByType("developer")
	assert.NoError(t, err)
	assert.Equal(t, 50.0, plan.Price)
	assert.Equal(t, 30, plan.ArticleRetentionDays)
	assert.Equal(t, 600, plan.MonthlyUploadLimit)

	// 测试企业版特性
	plan, err = service.GetPlanByType("enterprise")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, plan.Price)               // 联系销售
	assert.Equal(t, -1, plan.ArticleRetentionDays) // 无限期
	assert.Equal(t, -1, plan.MonthlyUploadLimit)   // 无限制
}

func TestPlanService_PlanComparison(t *testing.T) {
	service := NewPlanService()

	// 获取所有计划
	plans := service.GetAllPlans()
	assert.Len(t, plans, 5)

	// 验证计划按价格排序
	communityPlan, _ := service.GetPlanByType("community")
	developerPlan, _ := service.GetPlanByType("developer")
	proPlan, _ := service.GetPlanByType("pro")
	maxPlan, _ := service.GetPlanByType("max")

	assert.True(t, communityPlan.Price < developerPlan.Price)
	assert.True(t, developerPlan.Price < proPlan.Price)
	assert.True(t, proPlan.Price < maxPlan.Price)

	// 验证存储限制递增
	assert.True(t, communityPlan.StorageLimitMB < developerPlan.StorageLimitMB)
	assert.True(t, developerPlan.StorageLimitMB < proPlan.StorageLimitMB)
	assert.True(t, proPlan.StorageLimitMB < maxPlan.StorageLimitMB)
}

func TestPlanService_IsValidPlanType(t *testing.T) {
	service := NewPlanService()

	validTypes := []string{"community", "developer", "pro", "max", "enterprise"}
	invalidTypes := []string{"invalid", "test", "", "basic"}

	for _, planType := range validTypes {
		assert.True(t, service.IsValidPlanType(planType), "Plan type %s should be valid", planType)
	}

	for _, planType := range invalidTypes {
		assert.False(t, service.IsValidPlanType(planType), "Plan type %s should be invalid", planType)
	}
}
