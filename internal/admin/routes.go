package admin

import (
	"new-openclaw/internal/admin/handler"
	"new-openclaw/internal/admin/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册管理后台路由
func RegisterRoutes(r *gin.Engine) {
	admin := r.Group("/admin")
	{
		// 公开接口（无需认证）
		admin.POST("/login", handler.Login)

		// 需要认证的接口
		auth := admin.Group("")
		auth.Use(middleware.JWTAuth())
		{
			// 认证相关
			auth.POST("/logout", handler.Logout)
			auth.GET("/profile", handler.GetProfile)
			auth.POST("/refresh-token", handler.RefreshToken)

			// 仪表盘
			auth.GET("/dashboard", handler.Dashboard)

			// 管理员管理（仅超级管理员）
			admins := auth.Group("/admins")
			admins.Use(middleware.RequireRole("super_admin"))
			{
				admins.GET("", handler.ListAdmins)
				admins.POST("", handler.CreateAdmin)
				admins.PUT("/:id", handler.UpdateAdmin)
				admins.DELETE("/:id", handler.DeleteAdmin)
			}
		}
	}
}
