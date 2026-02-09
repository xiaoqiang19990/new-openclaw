package handler

import (
	"context"
	"net/http"
	"time"

	"new-openclaw/internal/database"

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
	status := gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "new-openclaw",
		"version":   "1.0.0",
	}

	// 检查 MySQL
	if db := database.GetMySQL(); db != nil {
		sqlDB, err := db.DB()
		if err == nil && sqlDB.Ping() == nil {
			status["mysql"] = "connected"
		} else {
			status["mysql"] = "disconnected"
		}
	} else {
		status["mysql"] = "not configured"
	}

	// 检查 Redis
	if rdb := database.GetRedis(); rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if rdb.Ping(ctx).Err() == nil {
			status["redis"] = "connected"
		} else {
			status["redis"] = "disconnected"
		}
	} else {
		status["redis"] = "not configured"
	}

	// 检查 MongoDB
	if mdb := database.GetMongoDB(); mdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if database.MongoClient.Ping(ctx, nil) == nil {
			status["mongodb"] = "connected"
		} else {
			status["mongodb"] = "disconnected"
		}
	} else {
		status["mongodb"] = "not configured"
	}

	c.JSON(http.StatusOK, status)
}
