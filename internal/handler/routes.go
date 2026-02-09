package handler

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/ping", Ping)
	r.GET("/health", HealthCheck)

	// API v1 分组
	v1 := r.Group("/api/v1")
	{
		// 用户相关
		v1.GET("/users", GetUsers)
		v1.GET("/users/:id", GetUserByID)
		v1.POST("/users", CreateUser)
		v1.PUT("/users/:id", UpdateUser)
		v1.DELETE("/users/:id", DeleteUser)
	}
}
