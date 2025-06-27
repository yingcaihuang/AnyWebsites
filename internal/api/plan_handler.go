package api

import (
	"net/http"
	"time"

	"anywebsites/internal/models"
	"anywebsites/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PlanHandler struct {
	planService *services.PlanService
}

func NewPlanHandler() *PlanHandler {
	return &PlanHandler{
		planService: services.NewPlanService(),
	}
}

// GetPlans 获取所有可用计划
func (h *PlanHandler) GetPlans(c *gin.Context) {
	plans := models.GetDefaultPlanConfigs()
	c.JSON(http.StatusOK, gin.H{
		"plans": plans,
	})
}

// GetUserPlan 获取用户当前计划
func (h *PlanHandler) GetUserPlan(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	subscription, err := h.planService.GetUserPlan(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscription": subscription,
	})
}

// GetUsageLimits 获取用户使用限制状态
func (h *PlanHandler) GetUsageLimits(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limitStatus, err := h.planService.CheckUsageLimits(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check usage limits"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"usage_limits": limitStatus,
	})
}

// UpgradePlan 升级用户计划
func (h *PlanHandler) UpgradePlan(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		PlanType      models.PlanType `json:"plan_type" binding:"required"`
		PaymentMethod string          `json:"payment_method"`
		Duration      int             `json:"duration"` // 订阅月数
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证计划类型
	validPlans := []models.PlanType{
		models.PlanDeveloper,
		models.PlanPro,
		models.PlanMax,
		models.PlanEnterprise,
	}

	isValidPlan := false
	for _, plan := range validPlans {
		if req.PlanType == plan {
			isValidPlan = true
			break
		}
	}

	if !isValidPlan {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan type"})
		return
	}

	// 计算过期时间
	var expiresAt *time.Time
	if req.PlanType != models.PlanEnterprise && req.Duration > 0 {
		expTime := time.Now().AddDate(0, req.Duration, 0)
		expiresAt = &expTime
	}

	// 升级计划
	err := h.planService.UpgradePlan(userID.(uuid.UUID), req.PlanType, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade plan: " + err.Error()})
		return
	}

	// 获取更新后的订阅信息
	subscription, err := h.planService.GetUserPlan(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Plan upgraded successfully",
		"subscription": subscription,
	})
}

// CancelPlan 取消用户计划（降级到免费版）
func (h *PlanHandler) CancelPlan(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	subscription, err := h.planService.DowngradeToFree(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel plan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Plan cancelled successfully, downgraded to Community Plan",
		"subscription": subscription,
	})
}

// GetPlanHistory 获取用户计划变更历史
func (h *PlanHandler) GetPlanHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var histories []models.PlanUpgradeHistory
	err := h.planService.GetPlanHistory(userID.(uuid.UUID), &histories)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get plan history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"histories": histories,
	})
}

// AdminGetUserPlan 管理员获取用户计划（管理员专用）
func (h *PlanHandler) AdminGetUserPlan(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	subscription, err := h.planService.GetUserPlan(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscription": subscription,
	})
}

// AdminUpgradeUserPlan 管理员升级用户计划
func (h *PlanHandler) AdminUpgradeUserPlan(c *gin.Context) {
	// 检查管理员权限
	isAdmin, exists := c.Get("is_admin")
	if !exists || !isAdmin.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		PlanType models.PlanType `json:"plan_type" binding:"required"`
		Duration int             `json:"duration"` // 订阅月数
		Reason   string          `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 计算过期时间
	var expiresAt *time.Time
	if req.PlanType != models.PlanEnterprise && req.Duration > 0 {
		expTime := time.Now().AddDate(0, req.Duration, 0)
		expiresAt = &expTime
	}

	// 升级计划
	err = h.planService.UpgradePlan(userID, req.PlanType, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade user plan: " + err.Error()})
		return
	}

	// 获取更新后的订阅信息
	subscription, err := h.planService.GetUserPlan(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated plan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "User plan upgraded successfully",
		"subscription": subscription,
	})
}
