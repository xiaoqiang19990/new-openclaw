package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"new-openclaw/internal/admin"
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

	// ========== 安全中间件配置 ==========

	// 1. 基础中间件
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())      // 请求 ID
	r.Use(middleware.SecureHeaders())  // 安全响应头

	// 2. CORS 跨域
	r.Use(middleware.Cors())

	// 3. IP 过滤（黑名单/白名单）
	ipFilterConfig := middleware.IPFilterConfig{
		WhitelistMode: cfg.Security.IPWhitelistMode,
		Whitelist:     cfg.Security.IPWhitelist,
		Blacklist:     cfg.Security.IPBlacklist,
		AllowPrivate:  true,
		TrustProxy:    true,
		ProxyHeader:   "X-Real-IP",
		BlockHandler:  middleware.DefaultIPFilterConfig.BlockHandler,
	}
	r.Use(middleware.IPFilterWithConfig(ipFilterConfig))

	// 4. 全局频率限制
	rateLimitConfig := middleware.RateLimitConfig{
		Window:       cfg.Security.RateLimitWindow,
		MaxRequests:  cfg.Security.RateLimitMaxRequests,
		KeyFunc:      middleware.DefaultRateLimitConfig.KeyFunc,
		LimitHandler: middleware.DefaultRateLimitConfig.LimitHandler,
	}
	r.Use(middleware.RateLimitWithConfig(rateLimitConfig))

	// 5. 请求日志审计
	auditConfig := middleware.AuditConfig{
		Enabled:             cfg.Security.AuditEnabled,
		Output:              cfg.Security.AuditOutput,
		FilePath:            cfg.Security.AuditFilePath,
		LogRequestBody:      true,
		LogResponseBody:     true,
		MaxRequestBodySize:  4096,
		MaxResponseBodySize: 4096,
		SensitiveFields:     []string{"password", "token", "secret", "key", "authorization"},
		ExcludePaths:        []string{"/ping", "/health", "/metrics"},
		Async:               true,
		BufferSize:          1000,
	}
	r.Use(middleware.AuditWithConfig(auditConfig))

	// 6. 安全审计（检测攻击行为）
	r.Use(middleware.SecurityAudit())

	// 7. 日志中间件
	r.Use(middleware.Logger())

	// 更新 JWT 配置
	middleware.DefaultJWTConfig = middleware.JWTConfig{
		SecretKey:     cfg.Security.JWTSecretKey,
		TokenExpiry:   cfg.Security.JWTExpiry,
		RefreshExpiry: cfg.Security.JWTRefreshExpiry,
		Issuer:        cfg.Security.JWTIssuer,
	}

	// 更新 API 签名配置
	middleware.DefaultSignatureConfig = middleware.SignatureConfig{
		SecretKey:      cfg.Security.APISignatureKey,
		Expiry:         cfg.Security.APISignatureExpiry,
		Algorithm:      "hmac-sha256",
		TimeTolerance:  time.Minute * 2,
		SignatureParam: "sign",
		TimestampParam: "timestamp",
		NonceParam:     "nonce",
		AppKeyParam:    "app_key",
		ValidateBody:   true,
	}

	// ========== 注册路由 ==========
	handler.RegisterRoutes(r)

	// 注册管理后台路由
	admin.RegisterRoutes(r)

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
	log.Printf("📋 安全功能已启用:")
	log.Printf("   - JWT Token 认证")
	log.Printf("   - 请求频率限制 (%d 次/%v)", cfg.Security.RateLimitMaxRequests, cfg.Security.RateLimitWindow)
	log.Printf("   - API 签名验证")
	log.Printf("   - IP 过滤 (白名单模式: %v)", cfg.Security.IPWhitelistMode)
	log.Printf("   - 请求日志审计 (输出: %s)", cfg.Security.AuditOutput)

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
