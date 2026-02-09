package main

import (
	"fmt"
	"log"
	"os"

	"new-openclaw/internal/handler"
	"new-openclaw/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// 获取端口，默认 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 设置运行模式
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)

	// 创建路由
	r := gin.New()

	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())

	// 注册路由
	handler.RegisterRoutes(r)

	// 启动服务
	addr := fmt.Sprintf(":%s", port)
	log.Printf("🚀 服务启动在 http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
