package handler

import (
	"net/http"

	"new-openclaw/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// Dashboard 管理后台首页
// @Summary 管理后台首页
// @Tags Admin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /admin/dashboard [get]
func Dashboard(c *gin.Context) {
	claims, _ := c.Get("admin_claims")
	adminClaims := claims.(*jwt.Claims)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"welcome":  "欢迎来到 OpenClaw 管理后台",
			"admin":    adminClaims.Username,
			"role":     adminClaims.Role,
			"menu": []gin.H{
				{"name": "仪表盘", "path": "/admin/dashboard", "icon": "dashboard"},
				{"name": "用户管理", "path": "/admin/users", "icon": "user"},
				{"name": "系统设置", "path": "/admin/settings", "icon": "setting"},
			},
			"stats": gin.H{
				"total_users":   0,
				"today_visits":  0,
				"total_orders":  0,
				"total_revenue": 0,
			},
		},
	})
}

// AdminIndex 管理后台HTML页面（可选）
func AdminIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/index.html", gin.H{
		"title": "OpenClaw 管理后台",
	})
}
