package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"new-openclaw/internal/database"
	"new-openclaw/internal/handler"
	"new-openclaw/internal/middleware"
	"new-openclaw/pkg/config"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 设置运行模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库连接
	if err := database.InitAll(cfg); err != nil {
		log.Printf("数据库初始化警告: %v", err)
	}

	// 优雅关闭
	defer database.CloseAll()

	// 创建路由
	r := gin.New()

	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())

	// 注册路由
	handler.RegisterRoutes(r)

	// 监听退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("正在关闭服务...")
		database.CloseAll()
		os.Exit(0)
	}()

	// 启动服务
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("🚀 服务启动在 http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
