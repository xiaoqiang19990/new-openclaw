package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig 频率限制配置
type RateLimitConfig struct {
	// 时间窗口
	Window time.Duration
	// 窗口内最大请求数
	MaxRequests int
	// 限制的 Key 生成函数（默认使用 IP）
	KeyFunc func(c *gin.Context) string
	// 被限制时的响应
	LimitHandler gin.HandlerFunc
}

// DefaultRateLimitConfig 默认频率限制配置
var DefaultRateLimitConfig = RateLimitConfig{
	Window:      time.Minute,
	MaxRequests: 60,
	KeyFunc: func(c *gin.Context) string {
		return c.ClientIP()
	},
	LimitHandler: func(c *gin.Context) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"code":    429,
			"message": "请求过于频繁，请稍后再试",
		})
		c.Abort()
	},
}

// rateLimitEntry 频率限制条目
type rateLimitEntry struct {
	count     int
	startTime time.Time
}

// RateLimiter 频率限制器
type RateLimiter struct {
	config  RateLimitConfig
	entries map[string]*rateLimitEntry
	mu      sync.RWMutex
}

// NewRateLimiter 创建频率限制器
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		config:  config,
		entries: make(map[string]*rateLimitEntry),
	}

	// 启动清理协程
	go rl.cleanup()

	return rl
}

// cleanup 定期清理过期条目
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.Window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.entries {
			if now.Sub(entry.startTime) > rl.config.Window {
				delete(rl.entries, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]

	if !exists || now.Sub(entry.startTime) > rl.config.Window {
		rl.entries[key] = &rateLimitEntry{
			count:     1,
			startTime: now,
		}
		return true
	}

	if entry.count >= rl.config.MaxRequests {
		return false
	}

	entry.count++
	return true
}

// GetRemaining 获取剩余请求数
func (rl *RateLimiter) GetRemaining(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	entry, exists := rl.entries[key]
	if !exists {
		return rl.config.MaxRequests
	}

	if time.Now().Sub(entry.startTime) > rl.config.Window {
		return rl.config.MaxRequests
	}

	remaining := rl.config.MaxRequests - entry.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RateLimit 频率限制中间件
func RateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(DefaultRateLimitConfig)
}

// RateLimitWithConfig 带配置的频率限制中间件
func RateLimitWithConfig(config RateLimitConfig) gin.HandlerFunc {
	limiter := NewRateLimiter(config)

	return func(c *gin.Context) {
		key := config.KeyFunc(c)

		if !limiter.Allow(key) {
			config.LimitHandler(c)
			return
		}

		// 添加响应头
		remaining := limiter.GetRemaining(key)
		c.Header("X-RateLimit-Limit", string(rune(config.MaxRequests)))
		c.Header("X-RateLimit-Remaining", string(rune(remaining)))

		c.Next()
	}
}

// APIRateLimit 针对 API 的频率限制（更严格）
func APIRateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	config := RateLimitConfig{
		Window:      window,
		MaxRequests: maxRequests,
		KeyFunc: func(c *gin.Context) string {
			// 优先使用用户 ID，否则使用 IP
			if userID, exists := c.Get("user_id"); exists {
				return "user:" + userID.(string)
			}
			return "ip:" + c.ClientIP()
		},
		LimitHandler: DefaultRateLimitConfig.LimitHandler,
	}

	return RateLimitWithConfig(config)
}

// EndpointRateLimit 针对特定端点的频率限制
func EndpointRateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	config := RateLimitConfig{
		Window:      window,
		MaxRequests: maxRequests,
		KeyFunc: func(c *gin.Context) string {
			// 使用 IP + 路径作为 Key
			return c.ClientIP() + ":" + c.FullPath()
		},
		LimitHandler: DefaultRateLimitConfig.LimitHandler,
	}

	return RateLimitWithConfig(config)
}

// SlidingWindowRateLimiter 滑动窗口频率限制器
type SlidingWindowRateLimiter struct {
	config    RateLimitConfig
	requests  map[string][]time.Time
	mu        sync.RWMutex
}

// NewSlidingWindowRateLimiter 创建滑动窗口频率限制器
func NewSlidingWindowRateLimiter(config RateLimitConfig) *SlidingWindowRateLimiter {
	rl := &SlidingWindowRateLimiter{
		config:   config,
		requests: make(map[string][]time.Time),
	}

	go rl.cleanup()

	return rl
}

// cleanup 清理过期请求记录
func (rl *SlidingWindowRateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.Window / 2)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, times := range rl.requests {
			var valid []time.Time
			for _, t := range times {
				if now.Sub(t) <= rl.config.Window {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = valid
			}
		}
		rl.mu.Unlock()
	}
}

// Allow 检查是否允许请求（滑动窗口）
func (rl *SlidingWindowRateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	times := rl.requests[key]

	// 过滤掉窗口外的请求
	var valid []time.Time
	for _, t := range times {
		if now.Sub(t) <= rl.config.Window {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.config.MaxRequests {
		rl.requests[key] = valid
		return false
	}

	valid = append(valid, now)
	rl.requests[key] = valid
	return true
}

// SlidingWindowRateLimit 滑动窗口频率限制中间件
func SlidingWindowRateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	config := RateLimitConfig{
		Window:       window,
		MaxRequests:  maxRequests,
		KeyFunc:      DefaultRateLimitConfig.KeyFunc,
		LimitHandler: DefaultRateLimitConfig.LimitHandler,
	}

	limiter := NewSlidingWindowRateLimiter(config)

	return func(c *gin.Context) {
		key := config.KeyFunc(c)

		if !limiter.Allow(key) {
			config.LimitHandler(c)
			return
		}

		c.Next()
	}
}
