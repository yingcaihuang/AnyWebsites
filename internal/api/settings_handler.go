package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"anywebsites/internal/database"
	"anywebsites/internal/models"
	"anywebsites/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SettingsHandler 设置处理器
type SettingsHandler struct {
	settingsService     *services.SettingsService
	configReloadService *services.ConfigReloadService
}

// NewSettingsHandler 创建设置处理器
func NewSettingsHandler(settingsService *services.SettingsService, configReloadService *services.ConfigReloadService) *SettingsHandler {
	return &SettingsHandler{
		settingsService:     settingsService,
		configReloadService: configReloadService,
	}
}

// SettingsPage 设置页面
func (h *SettingsHandler) SettingsPage(c *gin.Context) {
	username, _ := c.Get("username")

	// 获取所有分类
	categories, err := h.settingsService.GetCategories()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Title":   "错误",
			"Message": "加载设置分类失败: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":      "系统设置",
		"Page":       "settings",
		"Username":   username,
		"Categories": categories,
	})
}

// GetAllSettings 获取所有设置
func (h *SettingsHandler) GetAllSettings(c *gin.Context) {
	settings, err := h.settingsService.GetAllSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"settings": settings,
	})
}

// GetSettingsByCategory 按分类获取设置
func (h *SettingsHandler) GetSettingsByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "分类参数不能为空",
		})
		return
	}

	settings, err := h.settingsService.GetSettingsByCategory(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"settings": settings,
	})
}

// GetCategories 获取所有分类
func (h *SettingsHandler) GetCategories(c *gin.Context) {
	categories, err := h.settingsService.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"categories": categories,
	})
}

// CreateSetting 创建设置
func (h *SettingsHandler) CreateSetting(c *gin.Context) {
	var req models.SettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户ID
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "用户认证失败",
		})
		return
	}

	// 创建设置
	err = h.settingsService.SetSetting(
		req.Category,
		req.Key,
		req.Value,
		req.Description,
		userID,
		req.Reason,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 触发配置重载
	if h.configReloadService != nil {
		if err := h.configReloadService.TriggerReload(); err != nil {
			log.Printf("Config reload failed after setting creation: %v", err)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "设置创建成功",
	})
}

// UpdateSetting 更新设置
func (h *SettingsHandler) UpdateSetting(c *gin.Context) {
	settingID := c.Param("id")
	if settingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "设置ID不能为空",
		})
		return
	}

	var req models.SettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户ID
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "用户认证失败",
		})
		return
	}

	// 更新设置
	err = h.settingsService.SetSetting(
		req.Category,
		req.Key,
		req.Value,
		req.Description,
		userID,
		req.Reason,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 触发配置重载
	if h.configReloadService != nil {
		if err := h.configReloadService.TriggerReload(); err != nil {
			log.Printf("Config reload failed after setting update: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "设置更新成功",
	})
}

// DeleteSetting 删除设置
func (h *SettingsHandler) DeleteSetting(c *gin.Context) {
	settingID := c.Param("id")
	if settingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "设置ID不能为空",
		})
		return
	}

	// 解析请求体获取删除原因
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果解析失败，使用默认原因
		req.Reason = "通过API删除"
	}

	// 获取当前用户ID
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "用户认证失败",
		})
		return
	}

	// 首先需要根据ID获取设置的category和key
	// 这里需要添加一个通过ID获取设置的方法
	setting, err := h.getSettingByID(settingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "设置不存在",
		})
		return
	}

	// 删除设置
	err = h.settingsService.DeleteSetting(
		setting.Category,
		setting.Key,
		userID,
		req.Reason,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "设置删除成功",
	})
}

// GetSettingHistory 获取设置历史
func (h *SettingsHandler) GetSettingHistory(c *gin.Context) {
	category := c.Param("category")
	key := c.Param("key")

	if category == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "分类和键名不能为空",
		})
		return
	}

	// 获取限制数量
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	history, err := h.settingsService.GetSettingHistory(category, key, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"history": history,
	})
}

// ExportSettings 导出设置
func (h *SettingsHandler) ExportSettings(c *gin.Context) {
	backup, err := h.settingsService.ExportSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"backup":  backup,
	})
}

// ImportSettings 导入设置
func (h *SettingsHandler) ImportSettings(c *gin.Context) {
	var req struct {
		Backup    *models.SettingsBackup `json:"backup" binding:"required"`
		Overwrite bool                   `json:"overwrite"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户ID
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "用户认证失败",
		})
		return
	}

	// 导入设置
	err = h.settingsService.ImportSettings(req.Backup, userID, req.Overwrite)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "设置导入成功",
	})
}

// getCurrentUserID 获取当前用户ID
func (h *SettingsHandler) getCurrentUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}

	switch v := userIDStr.(type) {
	case string:
		return uuid.Parse(v)
	case uuid.UUID:
		return v, nil
	default:
		return uuid.Nil, fmt.Errorf("invalid user ID type")
	}
}

// getSettingByID 根据ID获取设置
func (h *SettingsHandler) getSettingByID(settingID string) (*models.SystemSetting, error) {
	id, err := uuid.Parse(settingID)
	if err != nil {
		return nil, fmt.Errorf("invalid setting ID: %w", err)
	}

	var setting models.SystemSetting
	err = database.DB.Where("id = ? AND is_active = ?", id, true).First(&setting).Error
	if err != nil {
		return nil, err
	}

	return &setting, nil
}

// ReloadConfig 手动重载配置
func (h *SettingsHandler) ReloadConfig(c *gin.Context) {
	if h.configReloadService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "配置重载服务未启用",
		})
		return
	}

	err := h.configReloadService.TriggerReload()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "配置重载失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置重载成功",
	})
}

// GetConfigReloadStatus 获取配置重载状态
func (h *SettingsHandler) GetConfigReloadStatus(c *gin.Context) {
	if h.configReloadService == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"status": gin.H{
				"enabled":  false,
				"handlers": []string{},
			},
		})
		return
	}

	handlers := h.configReloadService.GetReloadHandlers()
	currentConfig := h.configReloadService.GetCurrentConfig()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"status": gin.H{
			"enabled":  true,
			"handlers": handlers,
			"current_config": gin.H{
				"server": gin.H{
					"host": currentConfig.Server.Host,
					"port": currentConfig.Server.Port,
				},
				"database": gin.H{
					"host": currentConfig.Database.Host,
					"port": currentConfig.Database.Port,
					"name": currentConfig.Database.Name,
				},
			},
		},
	})
}
