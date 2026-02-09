package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	// JWT 配置
	JWT JWTConfig
	// 频率限制配置
	RateLimit RateLimitConfig
	// API 签名配置
	Signature SignatureConfig
	// IP 过滤配置
	IPFilter IPFilterConfig
	// 审计配置
	Audit AuditConfig
}

// DefaultSecurityConfig 默认安全配置
var DefaultSecurityConfig = SecurityConfig{
	JWT:       DefaultJWTConfig,
	RateLimit: DefaultRateLimitConfig,
	Signature: DefaultSignatureConfig,
	IPFilter:  DefaultIPFilterConfig,
	Audit:     DefaultAuditConfig,
}

// SecurityMiddleware 安全中间件组合
type SecurityMiddleware struct {
	config          SecurityConfig
	ipFilter        *DynamicIPFilter
	rateLimiter     *RateLimiter
	auditLogger     *AuditLogger
}

// NewSecurityMiddleware 创建安全中间件
func NewSecurityMiddleware(config SecurityConfig) *SecurityMiddleware {
	auditLogger, _ := NewAuditLogger(config.Audit)

	return &SecurityMiddleware{
		config:      config,
		ipFilter:    NewDynamicIPFilter(config.IPFilter),
		rateLimiter: NewRateLimiter(config.RateLimit),
		auditLogger: auditLogger,
	}
}

// Apply 应用所有安全中间件到路由组
func (sm *SecurityMiddleware) Apply(r *gin.RouterGroup) {
	// IP 过滤
	r.Use(sm.ipFilter.Middleware())
	// 频率限制
	r.Use(RateLimitWithConfig(sm.config.RateLimit))
	// 审计日志
	r.Use(AuditWithLogger(sm.auditLogger))
	// 安全审计（检测攻击）
	r.Use(SecurityAudit())
}

// ApplyJWT 应用 JWT 认证
func (sm *SecurityMiddleware) ApplyJWT(r *gin.RouterGroup) {
	r.Use(JWTAuthWithConfig(sm.config.JWT))
}

// ApplySignature 应用 API 签名验证
func (sm *SecurityMiddleware) ApplySignature(r *gin.RouterGroup) {
	r.Use(APISignatureWithConfig(sm.config.Signature))
}

// GetIPFilter 获取 IP 过滤器（用于动态管理）
func (sm *SecurityMiddleware) GetIPFilter() *DynamicIPFilter {
	return sm.ipFilter
}

// Close 关闭资源
func (sm *SecurityMiddleware) Close() {
	if sm.auditLogger != nil {
		sm.auditLogger.Close()
	}
}

// SecureHeaders 安全响应头中间件
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")
		// 防止 MIME 类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")
		// XSS 保护
		c.Header("X-XSS-Protection", "1; mode=block")
		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// 内容安全策略
		c.Header("Content-Security-Policy", "default-src 'self'")
		// HSTS（仅 HTTPS）
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// 权限策略
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// RequestID 请求 ID 中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Timeout 请求超时中间件
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置超时上下文
		// 注意：这里只是设置 header，实际超时控制需要在业务逻辑中处理
		c.Header("X-Request-Timeout", timeout.String())
		c.Next()
	}
}

// Recovery 增强的恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误
				c.Error(err.(error))
				
				c.JSON(500, gin.H{
					"code":    500,
					"message": "服务器内部错误",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// APIKeyAuth API Key 认证中间件
func APIKeyAuth(validKeys map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "缺少 API Key",
			})
			c.Abort()
			return
		}

		appName, valid := validKeys[apiKey]
		if !valid {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "无效的 API Key",
			})
			c.Abort()
			return
		}

		c.Set("app_name", appName)
		c.Next()
	}
}

// BasicAuth 基本认证中间件
func BasicAuth(accounts gin.Accounts) gin.HandlerFunc {
	return gin.BasicAuth(accounts)
}
