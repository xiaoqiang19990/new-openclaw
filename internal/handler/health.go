package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Ping 简单的 ping 接口
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

// HealthCheck 健康检查接口
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "new-openclaw",
		"version":   "1.0.0",
	})
}
