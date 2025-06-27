package api

import (
	"anywebsites/internal/auth"
	"anywebsites/internal/config"
	"anywebsites/internal/middleware"
	"anywebsites/internal/services"
	"html/template"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(cfg *config.Config, geoipService *services.GeoIPService) *gin.Engine {
	// 初始化 JWT
	auth.InitJWT(cfg)

	r := gin.Default()

	// 设置模板函数
	r.SetFuncMap(template.FuncMap{
		"add": func(a, b interface{}) int64 {
			var aVal, bVal int64
			switch v := a.(type) {
			case int:
				aVal = int64(v)
			case int64:
				aVal = v
			default:
				aVal = 0
			}
			switch v := b.(type) {
			case int:
				bVal = int64(v)
			case int64:
				bVal = v
			default:
				bVal = 0
			}
			return aVal + bVal
		},
		"sub": func(a, b interface{}) int64 {
			var aVal, bVal int64
			switch v := a.(type) {
			case int:
				aVal = int64(v)
			case int64:
				aVal = v
			default:
				aVal = 0
			}
			switch v := b.(type) {
			case int:
				bVal = int64(v)
			case int64:
				bVal = v
			default:
				bVal = 0
			}
			return aVal - bVal
		},
		"mul": func(a, b interface{}) int64 {
			var aVal, bVal int64
			switch v := a.(type) {
			case int:
				aVal = int64(v)
			case int64:
				aVal = v
			default:
				aVal = 0
			}
			switch v := b.(type) {
			case int:
				bVal = int64(v)
			case int64:
				bVal = v
			default:
				bVal = 0
			}
			return aVal * bVal
		},
		"div": func(a, b interface{}) int64 {
			var aVal, bVal int64
			switch v := a.(type) {
			case int:
				aVal = int64(v)
			case int64:
				aVal = v
			default:
				aVal = 0
			}
			switch v := b.(type) {
			case int:
				bVal = int64(v)
			case int64:
				bVal = v
			default:
				bVal = 0
			}
			if bVal == 0 {
				return 0
			}
			return aVal / bVal
		},
		"eq": func(a, b interface{}) bool { return a == b },
		"lt": func(a, b interface{}) bool {
			var aVal, bVal int64
			switch v := a.(type) {
			case int:
				aVal = int64(v)
			case int64:
				aVal = v
			default:
				return false
			}
			switch v := b.(type) {
			case int:
				bVal = int64(v)
			case int64:
				bVal = v
			default:
				return false
			}
			return aVal < bVal
		},
		"gt": func(a, b interface{}) bool {
			var aVal, bVal int64
			switch v := a.(type) {
			case int:
				aVal = int64(v)
			case int64:
				aVal = v
			default:
				return false
			}
			switch v := b.(type) {
			case int:
				bVal = int64(v)
			case int64:
				bVal = v
			default:
				return false
			}
			return aVal > bVal
		},
	})

	// 加载 HTML 模板
	r.LoadHTMLFiles(
		"web/templates/login.html",
		"web/templates/layout.html",
		"web/templates/dashboard.html",
		"web/templates/contents.html",
		"web/templates/content-form.html",
		"web/templates/users.html",
		"web/templates/user-form.html",
		"web/templates/user-plans.html",
		"web/templates/user-plan-edit.html",
		"web/templates/settings.html",
		"web/templates/analytics.html",
		"web/templates/geoip-monitor.html",
		"web/templates/plan-stats.html",
		"web/templates/error.html",
		"web/templates/admin/error.html",
	)

	// 根路径重定向到登录页面
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/admin/login")
	})

	// 静态文件
	r.Static("/static", "./web/static")

	// API 文档
	r.Static("/docs", "./docs")
	r.GET("/api-docs", func(c *gin.Context) {
		c.Redirect(302, "/docs/swagger-ui.html")
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 认证相关路由
	authHandler := NewAuthHandler()
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
	}

	// 内容相关路由
	contentHandler := NewContentHandler(geoipService)
	planHandler := NewPlanHandler()

	// 公开访问路由
	r.GET("/view/:id", contentHandler.View)

	// 公开 API 路由
	publicApiGroup := r.Group("/api/content")
	{
		// 公开上传（临时）
		publicApiGroup.POST("/upload", contentHandler.Upload)
	}

	// 需要认证的 API 路由
	authApiGroup := r.Group("/api/content")
	authApiGroup.Use(middleware.AuthMiddleware())
	{
		authApiGroup.GET("", contentHandler.List)          // 获取内容列表
		authApiGroup.GET("/:id", contentHandler.GetByID)   // 获取内容详情
		authApiGroup.PUT("/:id", contentHandler.Update)    // 更新内容
		authApiGroup.DELETE("/:id", contentHandler.Delete) // 删除内容
	}

	// 计划相关路由
	planApiGroup := r.Group("/api/plans")
	{
		planApiGroup.GET("", planHandler.GetPlans) // 获取所有计划（公开）
	}

	// 需要认证的计划路由
	authPlanGroup := r.Group("/api/plans")
	authPlanGroup.Use(middleware.AuthMiddleware())
	{
		authPlanGroup.GET("/current", planHandler.GetUserPlan)    // 获取当前用户计划
		authPlanGroup.GET("/usage", planHandler.GetUsageLimits)   // 获取使用限制
		authPlanGroup.POST("/upgrade", planHandler.UpgradePlan)   // 升级计划
		authPlanGroup.POST("/cancel", planHandler.CancelPlan)     // 取消计划
		authPlanGroup.GET("/history", planHandler.GetPlanHistory) // 获取计划历史
	}

	// 管理后台路由
	adminHandler := NewAdminHandler(geoipService)

	// 创建设置服务和处理器
	settingsService := services.NewSettingsService()

	// 创建配置重载服务
	configReloadService := services.NewConfigReloadService(settingsService, cfg)

	// 注册重载处理器
	configReloadService.RegisterReloadHandler(services.NewServerReloadHandler())
	configReloadService.RegisterReloadHandler(services.NewDatabaseReloadHandler())
	configReloadService.RegisterReloadHandler(services.NewSecurityReloadHandler())

	// 启动配置监听（每5分钟检查一次）
	configReloadService.StartWatching(5 * time.Minute)

	settingsHandler := NewSettingsHandler(settingsService, configReloadService)

	// 管理后台登录页面（无需认证）
	r.GET("/admin/login", adminHandler.LoginPage)
	r.POST("/admin/login", adminHandler.Login)

	// 需要认证的管理后台路由
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.AdminAuthMiddleware())
	{
		adminGroup.GET("", adminHandler.Dashboard)
		adminGroup.GET("/", adminHandler.Dashboard)
		adminGroup.GET("/logout", adminHandler.Logout)
		adminGroup.GET("/contents", adminHandler.Contents)
		adminGroup.GET("/contents/new", adminHandler.NewContent)
		adminGroup.POST("/contents/new", adminHandler.CreateContent)
		adminGroup.GET("/contents/:id/edit", adminHandler.EditContent)
		adminGroup.POST("/contents/:id/edit", adminHandler.UpdateContent)
		adminGroup.GET("/users", adminHandler.Users)
		adminGroup.GET("/users/new", adminHandler.NewUser)
		adminGroup.POST("/users/new", adminHandler.CreateUser)
		adminGroup.GET("/users/:id/edit", adminHandler.EditUser)
		adminGroup.POST("/users/:id/edit", adminHandler.UpdateUser)
		adminGroup.GET("/analytics", adminHandler.Analytics)
		adminGroup.GET("/geoip-monitor", adminHandler.GeoIPMonitor)
		adminGroup.GET("/settings", settingsHandler.SettingsPage)

		// 用户计划管理路由
		adminGroup.GET("/user-plans", adminHandler.UserPlans)
		adminGroup.GET("/user-plans/:id/edit", adminHandler.UserPlanEdit)
		adminGroup.POST("/user-plans/:id/update", adminHandler.UserPlanUpdate)
		adminGroup.POST("/user-plans/:id/upgrade", adminHandler.UpgradeUserPlan)
		adminGroup.POST("/user-plans/:id/downgrade", adminHandler.DowngradeUserPlan)
		adminGroup.GET("/plan-stats", adminHandler.PlanStats)

		// 管理后台 API
		adminApiGroup := adminGroup.Group("/api")
		{
			adminApiGroup.DELETE("/contents/:id", adminHandler.DeleteContent)
			adminApiGroup.POST("/contents/:id/restore", adminHandler.RestoreContent)
			adminApiGroup.POST("/contents/batch-delete", adminHandler.BatchDeleteContents)
			adminApiGroup.POST("/contents/batch-restore", adminHandler.BatchRestoreContents)
			adminApiGroup.POST("/users/:id/toggle-status", adminHandler.ToggleUserStatus)
			adminApiGroup.POST("/users/:id/toggle-admin", adminHandler.ToggleAdminStatus)
			adminApiGroup.POST("/users/:id/reset-api-key", adminHandler.ResetUserAPIKey)
			adminApiGroup.GET("/users/:id/details", adminHandler.GetUserDetails)
			adminApiGroup.POST("/users/:id/reset-password", adminHandler.ResetUserPassword)
			adminApiGroup.DELETE("/users/:id", adminHandler.DeleteUser)
			adminApiGroup.GET("/geoip-stats", adminHandler.GetGeoIPStats)

			// 设置管理 API
			adminApiGroup.GET("/settings", settingsHandler.GetAllSettings)
			adminApiGroup.GET("/settings/categories", settingsHandler.GetCategories)
			adminApiGroup.GET("/settings/category/:category", settingsHandler.GetSettingsByCategory)
			adminApiGroup.POST("/settings", settingsHandler.CreateSetting)
			adminApiGroup.PUT("/settings/:id", settingsHandler.UpdateSetting)
			adminApiGroup.DELETE("/settings/:id", settingsHandler.DeleteSetting)
			adminApiGroup.GET("/settings/:category/:key/history", settingsHandler.GetSettingHistory)
			adminApiGroup.GET("/settings/export", settingsHandler.ExportSettings)
			adminApiGroup.POST("/settings/import", settingsHandler.ImportSettings)
			adminApiGroup.POST("/settings/reload", settingsHandler.ReloadConfig)
			adminApiGroup.GET("/settings/reload-status", settingsHandler.GetConfigReloadStatus)
		}
	}

	return r
}
