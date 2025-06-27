package middleware

import (
	"net/http"

	"anywebsites/internal/auth"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware 管理后台认证中间件
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Cookie 获取 Token
		token, err := c.Cookie("admin_token")
		if err != nil || token == "" {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		// 验证 Token
		claims, err := auth.ValidateToken(token)
		if err != nil {
			c.SetCookie("admin_token", "", -1, "/admin", "", false, true)
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("is_admin", claims.IsAdmin)
		c.Next()
	}
}

// AdminOnlyMiddleware 仅管理员访问中间件
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.HTML(http.StatusForbidden, "error.html", gin.H{
				"Title":   "访问被拒绝",
				"Message": "您没有管理员权限",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
