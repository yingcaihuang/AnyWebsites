package api

import (
	"net/http"
	"strconv"
	"strings"

	"anywebsites/internal/database"
	"anywebsites/internal/models"
	"anywebsites/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ContentHandler struct {
	contentService *services.ContentService
}

func NewContentHandler(geoipService *services.GeoIPService) *ContentHandler {
	return &ContentHandler{
		contentService: services.NewContentService(geoipService),
	}
}

// Upload 上传 HTML 内容
func (h *ContentHandler) Upload(c *gin.Context) {
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 使用已注册的用户ID（临时解决方案）
	defaultUserID := uuid.MustParse("0b64c9d9-c45b-45eb-8324-36a855e34d69")

	content := &models.Content{
		UserID:      defaultUserID,
		Title:       req.Title,
		Content:     req.Content,
		ContentType: "text/html",
		IsActive:    true,
	}

	if err := database.DB.Create(content).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create content"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Content uploaded successfully",
		"id":      content.ID,
		"url":     "/view/" + content.ID.String(),
	})
}

// List 获取用户的内容列表
func (h *ContentHandler) List(c *gin.Context) {
	// 这里需要先添加 strconv 和 middleware 导入
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 获取分页参数
	page := 1
	limit := 10
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	contents, total, err := h.contentService.List(userID.(uuid.UUID), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get contents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contents": contents,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// GetByID 获取内容详情
func (h *ContentHandler) GetByID(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	content, err := h.contentService.GetByUserID(userID.(uuid.UUID), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": content})
}

// View 查看 HTML 内容
func (h *ContentHandler) View(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// 获取访问码（如果有）
	accessCode := c.Query("code")

	// 获取客户端IP地址（支持代理服务器传递的真实IP）
	clientIP := getRealClientIP(c)

	// 获取用户代理和来源
	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")

	// 使用 ContentService 的 ViewContent 方法来记录访问统计
	content, err := h.contentService.ViewContentWithAnalytics(id, accessCode, clientIP, userAgent, referer)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"Title":   "页面未找到",
			"Message": "请求的页面不存在或已过期",
		})
		return
	}

	// 返回 HTML 内容
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, content.Content)
}

// Update 更新内容
func (h *ContentHandler) Update(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req services.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	content, err := h.contentService.Update(userID.(uuid.UUID), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Content updated successfully",
		"content": content,
	})
}

// Delete 删除内容
func (h *ContentHandler) Delete(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = h.contentService.Delete(userID.(uuid.UUID), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Content deleted successfully"})
}

// getRealClientIP 获取真实的客户端IP地址，支持从代理服务器传递的真实IP
func getRealClientIP(c *gin.Context) string {
	// 1. 首先检查 X-Real-IP 头部（nginx proxy_set_header X-Real-IP $remote_addr;）
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	// 2. 检查 X-Forwarded-For 头部（可能包含多个IP，第一个是真实客户端IP）
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		// X-Forwarded-For 格式: client, proxy1, proxy2
		if firstIP := strings.Split(forwardedFor, ",")[0]; firstIP != "" {
			return strings.TrimSpace(firstIP)
		}
	}

	// 3. 检查 CF-Connecting-IP 头部（Cloudflare）
	if cfIP := c.GetHeader("CF-Connecting-IP"); cfIP != "" {
		return cfIP
	}

	// 4. 最后使用 Gin 的默认 ClientIP 方法
	return c.ClientIP()
}
