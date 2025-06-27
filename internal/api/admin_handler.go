package api

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"anywebsites/internal/auth"
	"anywebsites/internal/database"
	"anywebsites/internal/models"
	"anywebsites/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	geoipService *services.GeoIPService
	planService  *services.PlanService
}

func NewAdminHandler(geoipService *services.GeoIPService) *AdminHandler {
	return &AdminHandler{
		geoipService: geoipService,
		planService:  services.NewPlanService(),
	}
}

// LoginPage 显示登录页面
func (h *AdminHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"Title": "管理员登录",
	})
}

// Login 处理登录
func (h *AdminHandler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"Title":    "管理员登录",
			"Error":    "用户名和密码不能为空",
			"Username": username,
		})
		return
	}

	// 验证用户
	var user models.User
	if err := database.DB.Where("username = ? AND is_active = ?", username, true).First(&user).Error; err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"Title":    "管理员登录",
			"Error":    "用户名或密码错误",
			"Username": username,
		})
		return
	}

	if !auth.CheckPassword(password, user.Password) {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"Title":    "管理员登录",
			"Error":    "用户名或密码错误",
			"Username": username,
		})
		return
	}

	// 生成 JWT Token
	token, err := auth.GenerateToken(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"Title": "管理员登录",
			"Error": "登录失败，请重试",
		})
		return
	}

	// 设置 Cookie
	c.SetCookie("admin_token", token, 3600*24, "/admin", "", false, true)
	c.Redirect(http.StatusFound, "/admin")
}

// Logout 退出登录
func (h *AdminHandler) Logout(c *gin.Context) {
	c.SetCookie("admin_token", "", -1, "/admin", "", false, true)
	c.Redirect(http.StatusFound, "/admin/login")
}

// Dashboard 仪表板
func (h *AdminHandler) Dashboard(c *gin.Context) {
	// 获取统计数据
	var stats models.OverviewStats

	// 总内容数
	database.DB.Model(&models.Content{}).Count(&stats.TotalContents)

	// 活跃内容数
	database.DB.Model(&models.Content{}).Where("is_active = ?", true).Count(&stats.ActiveContents)

	// 总访问量（从 analytics 表获取实际访问记录数）
	database.DB.Model(&models.ContentAnalytics{}).Count(&stats.TotalViews)

	// 今日访问量（从 analytics 表获取今日访问记录数）
	today := time.Now().Format("2006-01-02")
	database.DB.Model(&models.ContentAnalytics{}).
		Where("DATE(access_time) = ?", today).
		Count(&stats.TodayViews)

	// 获取最近内容
	var recentContents []models.Content
	database.DB.Preload("User").Where("is_active = ?", true).
		Order("created_at DESC").
		Limit(5).
		Find(&recentContents)

	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":          "仪表板",
		"Page":           "dashboard",
		"Username":       username,
		"Stats":          stats,
		"RecentContents": recentContents,
	})
}

// Contents 内容管理页面
func (h *AdminHandler) Contents(c *gin.Context) {
	// 获取查询参数
	page := 1
	limit := 20
	search := c.Query("search")
	status := c.Query("status")
	userFilter := c.Query("user")

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	offset := (page - 1) * limit

	// 构建查询
	query := database.DB.Model(&models.Content{}).Preload("User")

	// 搜索条件
	if search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 状态筛选
	if status == "active" {
		query = query.Where("is_active = ?", true)
	} else if status == "inactive" {
		query = query.Where("is_active = ?", false)
	}

	// 用户筛选
	if userFilter != "" {
		if userID, err := uuid.Parse(userFilter); err == nil {
			query = query.Where("user_id = ?", userID)
		}
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取内容列表
	var contents []models.Content
	query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&contents)

	// 获取所有用户（用于筛选下拉框）
	var users []models.User
	database.DB.Where("is_active = ?", true).Find(&users)

	// 计算分页
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	var pageNumbers []int
	start := page - 2
	if start < 1 {
		start = 1
	}
	end := start + 4
	if end > totalPages {
		end = totalPages
		start = end - 4
		if start < 1 {
			start = 1
		}
	}
	for i := start; i <= end; i++ {
		pageNumbers = append(pageNumbers, i)
	}

	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":       "内容管理",
		"Page":        "contents",
		"Username":    username,
		"Contents":    contents,
		"Users":       users,
		"Total":       total,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"PageNumbers": pageNumbers,
		"Search":      search,
		"Status":      status,
		"UserFilter":  userFilter,
	})
}

// DeleteContent API 删除内容
func (h *AdminHandler) DeleteContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	result := database.DB.Model(&models.Content{}).
		Where("id = ?", id).
		Update("is_active", false)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Content not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// RestoreContent API 恢复内容
func (h *AdminHandler) RestoreContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	result := database.DB.Model(&models.Content{}).
		Where("id = ?", id).
		Update("is_active", true)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Content not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// BatchDeleteContents API 批量删除内容
func (h *AdminHandler) BatchDeleteContents(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No IDs provided"})
		return
	}

	// 转换字符串ID为UUID
	var uuids []uuid.UUID
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID format: " + idStr})
			return
		}
		uuids = append(uuids, id)
	}

	// 批量软删除
	result := database.DB.Model(&models.Content{}).
		Where("id IN ?", uuids).
		Update("is_active", false)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   result.RowsAffected,
	})
}

// BatchRestoreContents API 批量恢复内容
func (h *AdminHandler) BatchRestoreContents(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No IDs provided"})
		return
	}

	// 转换字符串ID为UUID
	var uuids []uuid.UUID
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID format: " + idStr})
			return
		}
		uuids = append(uuids, id)
	}

	// 批量恢复
	result := database.DB.Model(&models.Content{}).
		Where("id IN ?", uuids).
		Update("is_active", true)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   result.RowsAffected,
	})
}

// NewContent 新建内容页面
func (h *AdminHandler) NewContent(c *gin.Context) {
	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":    "新建内容",
		"Page":     "content-form",
		"Username": username,
		"IsEdit":   false,
	})
}

// CreateContent 创建内容
func (h *AdminHandler) CreateContent(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.HTML(http.StatusUnauthorized, "error.html", gin.H{
			"Title":   "未授权",
			"Message": "用户未登录",
		})
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")
	htmlContent := c.PostForm("html_content")

	if title == "" || htmlContent == "" {
		username, _ := c.Get("username")
		c.HTML(http.StatusBadRequest, "layout.html", gin.H{
			"Title":       "新建内容",
			"Page":        "content-form",
			"Username":    username,
			"IsEdit":      false,
			"Error":       "标题和内容不能为空",
			"Title_":      title,
			"Description": description,
			"HtmlContent": htmlContent,
		})
		return
	}

	content := models.Content{
		UserID:      userID.(uuid.UUID),
		Title:       title,
		Description: description,
		HTMLContent: htmlContent,
		IsPublic:    true,
		IsActive:    true,
	}

	if err := database.DB.Create(&content).Error; err != nil {
		username, _ := c.Get("username")
		c.HTML(http.StatusInternalServerError, "layout.html", gin.H{
			"Title":       "新建内容",
			"Page":        "content-form",
			"Username":    username,
			"IsEdit":      false,
			"Error":       "创建失败: " + err.Error(),
			"Title_":      title,
			"Description": description,
			"HtmlContent": htmlContent,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/contents")
}

// EditContent 编辑内容页面
func (h *AdminHandler) EditContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Title":   "错误",
			"Message": "无效的内容ID",
		})
		return
	}

	var content models.Content
	if err := database.DB.Preload("User").Where("id = ?", id).First(&content).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"Title":   "未找到",
			"Message": "内容不存在",
		})
		return
	}

	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":       "编辑内容",
		"Page":        "content-form",
		"Username":    username,
		"IsEdit":      true,
		"Content":     content,
		"Title_":      content.Title,
		"Description": content.Description,
		"HtmlContent": content.HTMLContent,
	})
}

// UpdateContent 更新内容
func (h *AdminHandler) UpdateContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Title":   "错误",
			"Message": "无效的内容ID",
		})
		return
	}

	var content models.Content
	if err := database.DB.Where("id = ?", id).First(&content).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"Title":   "未找到",
			"Message": "内容不存在",
		})
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")
	htmlContent := c.PostForm("html_content")

	if title == "" || htmlContent == "" {
		username, _ := c.Get("username")
		c.HTML(http.StatusBadRequest, "layout.html", gin.H{
			"Title":       "编辑内容",
			"Page":        "content-form",
			"Username":    username,
			"IsEdit":      true,
			"Content":     content,
			"Error":       "标题和内容不能为空",
			"Title_":      title,
			"Description": description,
			"HtmlContent": htmlContent,
		})
		return
	}

	content.Title = title
	content.Description = description
	content.HTMLContent = htmlContent

	if err := database.DB.Save(&content).Error; err != nil {
		username, _ := c.Get("username")
		c.HTML(http.StatusInternalServerError, "layout.html", gin.H{
			"Title":       "编辑内容",
			"Page":        "content-form",
			"Username":    username,
			"IsEdit":      true,
			"Content":     content,
			"Error":       "更新失败: " + err.Error(),
			"Title_":      title,
			"Description": description,
			"HtmlContent": htmlContent,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/contents")
}

// Users 用户管理页面
func (h *AdminHandler) Users(c *gin.Context) {
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	// 搜索参数
	search := c.Query("search")
	status := c.Query("status") // active, inactive, all

	// 构建查询
	query := database.DB.Model(&models.User{})

	// 搜索条件
	if search != "" {
		query = query.Where("username ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 状态筛选
	switch status {
	case "active":
		query = query.Where("is_active = ?", true)
	case "inactive":
		query = query.Where("is_active = ?", false)
	case "all":
		// 不添加条件，显示所有用户
	default:
		// 默认显示活跃用户
		query = query.Where("is_active = ?", true)
		status = "active"
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取用户列表
	var users []models.User
	query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&users)

	// 计算分页
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	var pageNumbers []int
	start := page - 2
	if start < 1 {
		start = 1
	}
	end := start + 4
	if end > totalPages {
		end = totalPages
		start = end - 4
		if start < 1 {
			start = 1
		}
	}
	for i := start; i <= end; i++ {
		pageNumbers = append(pageNumbers, i)
	}

	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":       "用户管理",
		"Page":        "users",
		"Username":    username,
		"Users":       users,
		"Total":       total,
		"CurrentPage": page,
		"TotalPages":  totalPages,
		"PageNumbers": pageNumbers,
		"Search":      search,
		"Status":      status,
	})
}

// ToggleUserStatus 切换用户状态
func (h *AdminHandler) ToggleUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	// 切换状态
	user.IsActive = !user.IsActive

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "is_active": user.IsActive})
}

// ToggleAdminStatus 切换管理员状态
func (h *AdminHandler) ToggleAdminStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	// 切换管理员状态
	user.IsAdmin = !user.IsAdmin

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "is_admin": user.IsAdmin})
}

// ResetUserAPIKey 重置用户API密钥
func (h *AdminHandler) ResetUserAPIKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	// 生成新的API密钥
	newAPIKey := uuid.New().String()[:32] // 只取前32个字符
	user.APIKey = newAPIKey

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "api_key": newAPIKey})
}

// NewUser 新建用户页面
func (h *AdminHandler) NewUser(c *gin.Context) {
	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":    "新建用户",
		"Page":     "user-form",
		"Username": username,
		"IsEdit":   false,
	})
}

// CreateUser 创建用户
func (h *AdminHandler) CreateUser(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	isAdmin := c.PostForm("is_admin") == "on"

	if username == "" || email == "" || password == "" {
		adminUsername, _ := c.Get("username")
		c.HTML(http.StatusBadRequest, "layout.html", gin.H{
			"Title":     "新建用户",
			"Page":      "user-form",
			"Username":  adminUsername,
			"IsEdit":    false,
			"Error":     "用户名、邮箱和密码不能为空",
			"Username_": username,
			"Email":     email,
		})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := database.DB.Where("username = ?", username).First(&existingUser).Error; err == nil {
		adminUsername, _ := c.Get("username")
		c.HTML(http.StatusBadRequest, "layout.html", gin.H{
			"Title":     "新建用户",
			"Page":      "user-form",
			"Username":  adminUsername,
			"IsEdit":    false,
			"Error":     "用户名已存在",
			"Username_": username,
			"Email":     email,
		})
		return
	}

	// 检查邮箱是否已存在
	if err := database.DB.Where("email = ?", email).First(&existingUser).Error; err == nil {
		adminUsername, _ := c.Get("username")
		c.HTML(http.StatusBadRequest, "layout.html", gin.H{
			"Title":     "新建用户",
			"Page":      "user-form",
			"Username":  adminUsername,
			"IsEdit":    false,
			"Error":     "邮箱已存在",
			"Username_": username,
			"Email":     email,
		})
		return
	}

	// 创建用户
	user := models.User{
		Username: username,
		Email:    email,
		IsActive: true,
		IsAdmin:  isAdmin,
		APIKey:   uuid.New().String()[:32], // 生成API密钥
	}

	// 加密密码
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		adminUsername, _ := c.Get("username")
		c.HTML(http.StatusInternalServerError, "layout.html", gin.H{
			"Title":     "新建用户",
			"Page":      "user-form",
			"Username":  adminUsername,
			"IsEdit":    false,
			"Error":     "密码加密失败: " + err.Error(),
			"Username_": username,
			"Email":     email,
		})
		return
	}
	user.Password = hashedPassword

	if err := database.DB.Create(&user).Error; err != nil {
		adminUsername, _ := c.Get("username")
		c.HTML(http.StatusInternalServerError, "layout.html", gin.H{
			"Title":     "新建用户",
			"Page":      "user-form",
			"Username":  adminUsername,
			"IsEdit":    false,
			"Error":     "创建用户失败: " + err.Error(),
			"Username_": username,
			"Email":     email,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/users")
}

// GetUserDetails 获取用户详情
func (h *AdminHandler) GetUserDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	// 获取用户的内容统计
	var contentCount int64
	database.DB.Model(&models.Content{}).Where("user_id = ?", id).Count(&contentCount)

	// 获取用户的活跃内容统计
	var activeContentCount int64
	database.DB.Model(&models.Content{}).Where("user_id = ? AND is_active = ?", id, true).Count(&activeContentCount)

	// 获取用户最近的内容
	var recentContents []models.Content
	database.DB.Where("user_id = ?", id).Order("created_at DESC").Limit(5).Find(&recentContents)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"api_key":    user.APIKey,
			"is_active":  user.IsActive,
			"is_admin":   user.IsAdmin,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
		"stats": gin.H{
			"total_contents":  contentCount,
			"active_contents": activeContentCount,
		},
		"recent_contents": recentContents,
	})
}

// EditUser 编辑用户页面
func (h *AdminHandler) EditUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":    "编辑用户",
		"Page":     "user-form",
		"Username": username,
		"IsEdit":   true,
		"User":     user,
		"Email":    user.Email,
		"IsAdmin":  user.IsAdmin,
	})
}

// UpdateUser 更新用户
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	email := c.PostForm("email")
	isAdmin := c.PostForm("is_admin") == "on"

	if email == "" {
		adminUsername, _ := c.Get("username")
		c.HTML(http.StatusBadRequest, "layout.html", gin.H{
			"Title":    "编辑用户",
			"Page":     "user-form",
			"Username": adminUsername,
			"IsEdit":   true,
			"User":     user,
			"Error":    "邮箱不能为空",
			"Email":    email,
			"IsAdmin":  isAdmin,
		})
		return
	}

	// 检查邮箱是否已被其他用户使用
	var existingUser models.User
	if err := database.DB.Where("email = ? AND id != ?", email, id).First(&existingUser).Error; err == nil {
		adminUsername, _ := c.Get("username")
		c.HTML(http.StatusBadRequest, "layout.html", gin.H{
			"Title":    "编辑用户",
			"Page":     "user-form",
			"Username": adminUsername,
			"IsEdit":   true,
			"User":     user,
			"Error":    "邮箱已被其他用户使用",
			"Email":    email,
			"IsAdmin":  isAdmin,
		})
		return
	}

	// 更新用户信息
	user.Email = email
	user.IsAdmin = isAdmin

	if err := database.DB.Save(&user).Error; err != nil {
		adminUsername, _ := c.Get("username")
		c.HTML(http.StatusInternalServerError, "layout.html", gin.H{
			"Title":    "编辑用户",
			"Page":     "user-form",
			"Username": adminUsername,
			"IsEdit":   true,
			"User":     user,
			"Error":    "更新用户失败: " + err.Error(),
			"Email":    email,
			"IsAdmin":  isAdmin,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/users")
}

// ResetUserPassword 重置用户密码
func (h *AdminHandler) ResetUserPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	// 生成随机密码
	newPassword := generateRandomPassword(12)

	// 加密密码
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Password encryption failed"})
		return
	}

	// 更新密码
	user.Password = hashedPassword
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 在实际应用中，这里应该通过邮件发送新密码
	// 现在直接返回新密码（仅用于演示）
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"new_password": newPassword,
		"message":      "密码重置成功，新密码已生成",
	})
}

// DeleteUser 删除用户
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "User not found"})
		return
	}

	// 检查是否是最后一个管理员
	if user.IsAdmin {
		var adminCount int64
		database.DB.Model(&models.User{}).Where("is_admin = ? AND is_active = ?", true, true).Count(&adminCount)
		if adminCount <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "不能删除最后一个管理员账户",
			})
			return
		}
	}

	// 开始事务
	tx := database.DB.Begin()

	// 删除用户的所有内容
	if err := tx.Where("user_id = ?", id).Delete(&models.Content{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete user contents"})
		return
	}

	// 删除用户的所有分析数据
	if err := tx.Where("user_id = ?", id).Delete(&models.ContentAnalytics{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete user analytics"})
		return
	}

	// 删除用户
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 提交事务
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户及其所有数据已删除",
	})
}

// generateRandomPassword 生成随机密码
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}

// Analytics 统计分析页面
func (h *AdminHandler) Analytics(c *gin.Context) {
	// 获取时间范围参数
	timeRange := c.DefaultQuery("range", "7d") // 默认7天

	// 获取总览统计
	overviewStats := h.getOverviewStats()

	// 获取流量趋势数据
	trafficStats := h.getTrafficStats(timeRange)

	// 获取地理位置统计
	geoStats := h.getGeoStats(timeRange)

	// 获取国家分布统计
	countryStats := h.getCountryStats(timeRange)

	// 获取来源统计
	refererStats := h.getRefererStats(timeRange)

	// 获取热门内容
	popularContents := h.getPopularContents(timeRange)

	// 获取用户活跃度统计
	userActivityStats := h.getUserActivityStats(timeRange)

	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":             "统计分析",
		"Page":              "analytics",
		"Username":          username,
		"TimeRange":         timeRange,
		"OverviewStats":     overviewStats,
		"TrafficStats":      trafficStats,
		"GeoStats":          geoStats,
		"CountryStats":      countryStats,
		"RefererStats":      refererStats,
		"PopularContents":   popularContents,
		"UserActivityStats": userActivityStats,
	})
}

// getOverviewStats 获取总览统计
func (h *AdminHandler) getOverviewStats() models.OverviewStats {
	var stats models.OverviewStats

	// 总内容数
	database.DB.Model(&models.Content{}).Count(&stats.TotalContents)

	// 活跃内容数
	database.DB.Model(&models.Content{}).Where("is_active = ?", true).Count(&stats.ActiveContents)

	// 总访问量
	database.DB.Model(&models.ContentAnalytics{}).Count(&stats.TotalViews)

	// 今日访问量
	today := time.Now().Format("2006-01-02")
	database.DB.Model(&models.ContentAnalytics{}).
		Where("DATE(access_time) = ?", today).
		Count(&stats.TodayViews)

	// 独立访客数（基于IP地址）
	database.DB.Model(&models.ContentAnalytics{}).
		Distinct("ip_address").
		Count(&stats.UniqueVisitors)

	return stats
}

// getTrafficStats 获取流量趋势统计
func (h *AdminHandler) getTrafficStats(timeRange string) []models.TrafficStats {
	var stats []models.TrafficStats
	var days int

	switch timeRange {
	case "1d":
		days = 1
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		days = 7
	}

	// 生成日期范围
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")

		var views int64
		var uniqueIPs int64

		// 获取当日访问量
		database.DB.Model(&models.ContentAnalytics{}).
			Where("DATE(access_time) = ?", date).
			Count(&views)

		// 获取当日独立IP数
		database.DB.Model(&models.ContentAnalytics{}).
			Where("DATE(access_time) = ?", date).
			Distinct("ip_address").
			Count(&uniqueIPs)

		stats = append(stats, models.TrafficStats{
			Date:      date,
			Views:     views,
			UniqueIPs: uniqueIPs,
		})
	}

	return stats
}

// getGeoStats 获取地理位置统计
func (h *AdminHandler) getGeoStats(timeRange string) []models.GeoStats {
	var stats []models.GeoStats
	var days int

	switch timeRange {
	case "1d":
		days = 1
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		days = 7
	}

	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	database.DB.Model(&models.ContentAnalytics{}).
		Select("country, region, city, COUNT(*) as count").
		Where("DATE(access_time) >= ? AND country != ''", startDate).
		Group("country, region, city").
		Order("count DESC").
		Limit(20).
		Find(&stats)

	return stats
}

// getCountryStats 获取国家分布统计
func (h *AdminHandler) getCountryStats(timeRange string) []models.CountryStats {
	var stats []models.CountryStats
	var days int

	switch timeRange {
	case "1d":
		days = 1
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		days = 7
	}

	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	database.DB.Model(&models.ContentAnalytics{}).
		Select("country, COUNT(*) as count").
		Where("DATE(access_time) >= ? AND country != ''", startDate).
		Group("country").
		Order("count DESC").
		Limit(10).
		Find(&stats)

	return stats
}

// getRefererStats 获取来源统计
func (h *AdminHandler) getRefererStats(timeRange string) []models.RefererStats {
	var stats []models.RefererStats
	var days int

	switch timeRange {
	case "1d":
		days = 1
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		days = 7
	}

	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	database.DB.Model(&models.ContentAnalytics{}).
		Select("referer, COUNT(*) as count").
		Where("DATE(access_time) >= ? AND referer != ''", startDate).
		Group("referer").
		Order("count DESC").
		Limit(10).
		Find(&stats)

	return stats
}

// getPopularContents 获取热门内容
func (h *AdminHandler) getPopularContents(timeRange string) []struct {
	models.Content
	RecentViews int64 `json:"recent_views"`
} {
	var contents []struct {
		models.Content
		RecentViews int64 `json:"recent_views"`
	}

	var days int
	switch timeRange {
	case "1d":
		days = 1
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		days = 7
	}

	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	database.DB.Table("contents").
		Select("contents.*, COUNT(content_analytics.id) as recent_views").
		Joins("LEFT JOIN content_analytics ON contents.id = content_analytics.content_id AND DATE(content_analytics.access_time) >= ?", startDate).
		Where("contents.is_active = ?", true).
		Group("contents.id").
		Order("recent_views DESC").
		Limit(10).
		Find(&contents)

	return contents
}

// getUserActivityStats 获取用户活跃度统计
func (h *AdminHandler) getUserActivityStats(timeRange string) []struct {
	Username     string `json:"username"`
	ContentCount int64  `json:"content_count"`
	ViewCount    int64  `json:"view_count"`
} {
	var stats []struct {
		Username     string `json:"username"`
		ContentCount int64  `json:"content_count"`
		ViewCount    int64  `json:"view_count"`
	}

	var days int
	switch timeRange {
	case "1d":
		days = 1
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		days = 7
	}

	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	database.DB.Table("users").
		Select("users.username, COUNT(DISTINCT contents.id) as content_count, COUNT(content_analytics.id) as view_count").
		Joins("LEFT JOIN contents ON users.id = contents.user_id").
		Joins("LEFT JOIN content_analytics ON contents.id = content_analytics.content_id AND DATE(content_analytics.access_time) >= ?", startDate).
		Where("users.is_active = ?", true).
		Group("users.id, users.username").
		Order("view_count DESC").
		Limit(10).
		Find(&stats)

	return stats
}

// GeoIPMonitor GeoIP 服务监控页面
func (h *AdminHandler) GeoIPMonitor(c *gin.Context) {
	username, _ := c.Get("username")

	// 获取 GeoIP 服务统计信息
	var serviceStats map[string]interface{}
	var cacheStats map[string]interface{}

	if h.geoipService != nil {
		serviceStats = h.geoipService.GetServiceStats()
		cacheStats = h.geoipService.GetCacheStats()
	} else {
		serviceStats = map[string]interface{}{
			"total_requests":   0,
			"cache_hits":       0,
			"cache_misses":     0,
			"cache_hit_rate":   "0.00%",
			"batch_processed":  0,
			"direct_processed": 0,
			"errors":           0,
			"last_error":       "",
			"last_error_time":  "",
		}
		cacheStats = map[string]interface{}{
			"cache_size":    0,
			"cache_expiry":  "1h0m0s",
			"cache_entries": 0,
		}
	}

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":        "GeoIP 服务监控",
		"Page":         "geoip-monitor",
		"Username":     username,
		"ServiceStats": serviceStats,
		"CacheStats":   cacheStats,
	})
}

// GetGeoIPStats 获取 GeoIP 服务统计信息 API
func (h *AdminHandler) GetGeoIPStats(c *gin.Context) {
	if h.geoipService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "GeoIP service not available",
		})
		return
	}

	serviceStats := h.geoipService.GetServiceStats()
	cacheStats := h.geoipService.GetCacheStats()

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"service_stats": serviceStats,
		"cache_stats":   cacheStats,
		"timestamp":     time.Now().Format("2006-01-02 15:04:05"),
	})
}

// UserPlans 用户计划管理页面
func (h *AdminHandler) UserPlans(c *gin.Context) {
	var users []models.User
	if err := database.DB.Preload("Subscription").Find(&users).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "admin/error.html", gin.H{
			"title": "Error",
			"error": "Failed to load users",
		})
		return
	}

	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":    "用户计划管理",
		"Page":     "user-plans",
		"Username": username,
		"Users":    users,
	})
}

// UserPlanEdit 编辑用户计划页面
func (h *AdminHandler) UserPlanEdit(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := database.DB.Preload("Subscription").First(&user, "id = ?", userID).Error; err != nil {
		c.HTML(http.StatusNotFound, "admin/error.html", gin.H{
			"title": "Error",
			"error": "User not found",
		})
		return
	}

	// 获取所有计划
	plans := models.GetDefaultPlanConfigs()

	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":    "编辑用户计划",
		"Page":     "user-plan-edit",
		"Username": username,
		"User":     user,
		"Plans":    plans,
	})
}

// UserPlanUpdate 更新用户计划
func (h *AdminHandler) UserPlanUpdate(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		PlanType string `form:"plan_type"`
		Duration int    `form:"duration"`
		Reason   string `form:"reason"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "admin/error.html", gin.H{
			"title": "Error",
			"error": "Invalid form data",
		})
		return
	}

	// 解析用户ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "admin/error.html", gin.H{
			"title": "Error",
			"error": "Invalid user ID",
		})
		return
	}

	// 计算过期时间
	var expiresAt *time.Time
	if req.PlanType != "enterprise" && req.Duration > 0 {
		expTime := time.Now().AddDate(0, req.Duration, 0)
		expiresAt = &expTime
	}

	// 升级计划
	planType := models.PlanType(req.PlanType)
	if err := h.planService.UpgradePlan(userUUID, planType, expiresAt); err != nil {
		c.HTML(http.StatusInternalServerError, "admin/error.html", gin.H{
			"title": "Error",
			"error": "Failed to update user plan: " + err.Error(),
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/user-plans")
}

// PlanStats 计划统计页面
func (h *AdminHandler) PlanStats(c *gin.Context) {
	// 统计各计划的用户数
	var stats []struct {
		PlanType  string `json:"plan_type"`
		UserCount int64  `json:"user_count"`
	}

	err := database.DB.Table("user_subscriptions").
		Select("plan_type, COUNT(*) as user_count").
		Where("status = ?", models.StatusActive).
		Group("plan_type").
		Scan(&stats).Error

	if err != nil {
		c.HTML(http.StatusInternalServerError, "admin/error.html", gin.H{
			"title": "Error",
			"error": "Failed to load plan statistics",
		})
		return
	}

	// 统计总收入（模拟）
	var totalRevenue float64
	for _, stat := range stats {
		switch stat.PlanType {
		case "developer":
			totalRevenue += float64(stat.UserCount) * 50.0
		case "pro":
			totalRevenue += float64(stat.UserCount) * 100.0
		case "max":
			totalRevenue += float64(stat.UserCount) * 250.0
		}
	}

	username, _ := c.Get("username")

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":        "计划统计",
		"Page":         "plan-stats",
		"Username":     username,
		"Stats":        stats,
		"TotalRevenue": totalRevenue,
	})
}

// UpgradeUserPlan 升级用户计划
func (h *AdminHandler) UpgradeUserPlan(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	var req struct {
		PlanType  string     `json:"plan_type" binding:"required"`
		ExpiresAt *time.Time `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 创建用户服务实例
	userService := services.NewUserService(database.DB)

	// 升级用户计划
	if err := userService.UpgradeUserPlan(userID, req.PlanType, req.ExpiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "升级失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户计划升级成功",
	})
}

// DowngradeUserPlan 降级用户计划
func (h *AdminHandler) DowngradeUserPlan(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	var req struct {
		PlanType string `json:"plan_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 创建用户服务实例
	userService := services.NewUserService(database.DB)

	// 降级用户计划
	if err := userService.DowngradeUserPlan(userID, req.PlanType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "降级失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户计划降级成功",
	})
}
