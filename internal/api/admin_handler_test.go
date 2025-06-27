package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"anywebsites/internal/models"
	"anywebsites/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestRouter() (*gin.Engine, *gorm.DB) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移表结构
	db.AutoMigrate(&models.User{}, &models.UserSubscription{}, &models.PlanConfig{})

	// 创建路由
	router := gin.New()

	// 创建处理器
	planService := services.NewPlanService()
	adminHandler := &AdminHandler{
		planService: planService,
	}

	// 设置路由
	api := router.Group("/admin")
	{
		api.POST("/user-plans/:id/upgrade", adminHandler.UpgradeUserPlan)
		api.POST("/user-plans/:id/downgrade", adminHandler.DowngradeUserPlan)
		api.GET("/plan-stats", adminHandler.PlanStats)
	}

	return router, db
}

func TestAdminHandler_UpgradeUserPlan(t *testing.T) {
	router, db := setupTestRouter()

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

	// 准备请求数据
	requestData := map[string]interface{}{
		"plan_type":  "developer",
		"expires_at": time.Now().AddDate(0, 1, 0).Format(time.RFC3339),
	}
	jsonData, _ := json.Marshal(requestData)

	// 创建请求
	req, _ := http.NewRequest("POST", "/admin/user-plans/"+userID.String()+"/upgrade", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "用户计划升级成功", response["message"])
}

func TestAdminHandler_UpgradeUserPlan_InvalidUserID(t *testing.T) {
	router, _ := setupTestRouter()

	// 准备请求数据
	requestData := map[string]interface{}{
		"plan_type": "developer",
	}
	jsonData, _ := json.Marshal(requestData)

	// 创建请求（使用无效的用户ID）
	req, _ := http.NewRequest("POST", "/admin/user-plans/invalid-uuid/upgrade", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "升级失败")
}

func TestAdminHandler_UpgradeUserPlan_MissingPlanType(t *testing.T) {
	router, _ := setupTestRouter()

	userID := uuid.New()

	// 准备请求数据（缺少 plan_type）
	requestData := map[string]interface{}{
		"expires_at": time.Now().AddDate(0, 1, 0).Format(time.RFC3339),
	}
	jsonData, _ := json.Marshal(requestData)

	// 创建请求
	req, _ := http.NewRequest("POST", "/admin/user-plans/"+userID.String()+"/upgrade", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "请求参数无效")
}

func TestAdminHandler_DowngradeUserPlan(t *testing.T) {
	router, db := setupTestRouter()

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

	// 创建现有订阅
	subscription := models.UserSubscription{
		UserID:   userID,
		PlanType: models.PlanType("pro"),
		Status:   "active",
	}
	db.Create(&subscription)

	// 准备请求数据
	requestData := map[string]interface{}{
		"plan_type": "developer",
	}
	jsonData, _ := json.Marshal(requestData)

	// 创建请求
	req, _ := http.NewRequest("POST", "/admin/user-plans/"+userID.String()+"/downgrade", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "用户计划降级成功", response["message"])
}

func TestAdminHandler_DowngradeUserPlan_EmptyUserID(t *testing.T) {
	router, _ := setupTestRouter()

	// 准备请求数据
	requestData := map[string]interface{}{
		"plan_type": "community",
	}
	jsonData, _ := json.Marshal(requestData)

	// 创建请求（空的用户ID）
	req, _ := http.NewRequest("POST", "/admin/user-plans//downgrade", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应 - 应该是404因为路由不匹配
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdminHandler_DowngradeUserPlan_InvalidJSON(t *testing.T) {
	router, _ := setupTestRouter()

	userID := uuid.New()

	// 创建请求（无效的JSON）
	req, _ := http.NewRequest("POST", "/admin/user-plans/"+userID.String()+"/downgrade", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "请求参数无效")
}

func TestAdminHandler_PlanStats(t *testing.T) {
	router, db := setupTestRouter()

	// 创建测试用户和订阅
	users := []models.User{
		{ID: uuid.New(), Username: "user1", Email: "user1@example.com", APIKey: "key1", IsActive: true},
		{ID: uuid.New(), Username: "user2", Email: "user2@example.com", APIKey: "key2", IsActive: true},
		{ID: uuid.New(), Username: "user3", Email: "user3@example.com", APIKey: "key3", IsActive: true},
	}

	for _, user := range users {
		db.Create(&user)
	}

	// 创建订阅
	subscriptions := []models.UserSubscription{
		{UserID: users[0].ID, PlanType: models.PlanType("developer"), Status: models.StatusActive},
		{UserID: users[1].ID, PlanType: models.PlanType("developer"), Status: models.StatusActive},
		{UserID: users[2].ID, PlanType: models.PlanType("pro"), Status: models.StatusActive},
	}

	for _, sub := range subscriptions {
		db.Create(&sub)
	}

	// 创建请求
	req, _ := http.NewRequest("GET", "/admin/plan-stats", nil)

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	// 验证响应是HTML（因为这是一个页面渲染端点）
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
}

func TestAdminHandler_UpgradeUserPlan_WithoutExpiresAt(t *testing.T) {
	router, db := setupTestRouter()

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

	// 准备请求数据（不包含过期时间）
	requestData := map[string]interface{}{
		"plan_type": "enterprise",
	}
	jsonData, _ := json.Marshal(requestData)

	// 创建请求
	req, _ := http.NewRequest("POST", "/admin/user-plans/"+userID.String()+"/upgrade", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "用户计划升级成功", response["message"])
}
