package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SignatureConfig API 签名配置
type SignatureConfig struct {
	// 签名密钥
	SecretKey string
	// 签名有效期（防止重放攻击）
	Expiry time.Duration
	// 签名算法：hmac-sha256, md5
	Algorithm string
	// 时间戳容差（允许的时间偏差）
	TimeTolerance time.Duration
	// 签名参数名
	SignatureParam string
	// 时间戳参数名
	TimestampParam string
	// Nonce 参数名（防重放）
	NonceParam string
	// AppKey 参数名
	AppKeyParam string
	// 是否验证 Body
	ValidateBody bool
}

// DefaultSignatureConfig 默认签名配置
var DefaultSignatureConfig = SignatureConfig{
	SecretKey:      "your-api-secret-key",
	Expiry:         time.Minute * 5,
	Algorithm:      "hmac-sha256",
	TimeTolerance:  time.Minute * 2,
	SignatureParam: "sign",
	TimestampParam: "timestamp",
	NonceParam:     "nonce",
	AppKeyParam:    "app_key",
	ValidateBody:   true,
}

// nonceStore 用于存储已使用的 nonce（防重放）
var nonceStore = make(map[string]time.Time)

// APISignature API 签名验证中间件
func APISignature() gin.HandlerFunc {
	return APISignatureWithConfig(DefaultSignatureConfig)
}

// APISignatureWithConfig 带配置的 API 签名验证中间件
func APISignatureWithConfig(config SignatureConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取签名参数
		signature := c.GetHeader("X-Signature")
		if signature == "" {
			signature = c.Query(config.SignatureParam)
		}

		timestamp := c.GetHeader("X-Timestamp")
		if timestamp == "" {
			timestamp = c.Query(config.TimestampParam)
		}

		nonce := c.GetHeader("X-Nonce")
		if nonce == "" {
			nonce = c.Query(config.NonceParam)
		}

		appKey := c.GetHeader("X-App-Key")
		if appKey == "" {
			appKey = c.Query(config.AppKeyParam)
		}

		// 验证必要参数
		if signature == "" || timestamp == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "缺少签名参数",
			})
			c.Abort()
			return
		}

		// 验证时间戳
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的时间戳",
			})
			c.Abort()
			return
		}

		requestTime := time.Unix(ts, 0)
		now := time.Now()

		// 检查时间戳是否在有效范围内
		if now.Sub(requestTime) > config.Expiry || requestTime.Sub(now) > config.TimeTolerance {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求已过期",
			})
			c.Abort()
			return
		}

		// 检查 nonce 是否已使用（防重放攻击）
		if nonce != "" {
			if _, exists := nonceStore[nonce]; exists {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    400,
					"message": "重复的请求",
				})
				c.Abort()
				return
			}
			nonceStore[nonce] = now
			// 清理过期的 nonce
			go cleanupNonce(config.Expiry)
		}

		// 构建签名字符串
		signString := buildSignString(c, config, timestamp, nonce, appKey)

		// 计算签名
		expectedSign := calculateSignature(signString, config.SecretKey, config.Algorithm)

		// 验证签名
		if !hmac.Equal([]byte(signature), []byte(expectedSign)) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "签名验证失败",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// buildSignString 构建签名字符串
func buildSignString(c *gin.Context, config SignatureConfig, timestamp, nonce, appKey string) string {
	var parts []string

	// 添加请求方法
	parts = append(parts, c.Request.Method)

	// 添加请求路径
	parts = append(parts, c.Request.URL.Path)

	// 添加排序后的查询参数
	queryParams := c.Request.URL.Query()
	var queryKeys []string
	for key := range queryParams {
		// 排除签名相关参数
		if key != config.SignatureParam && key != config.TimestampParam && 
		   key != config.NonceParam && key != config.AppKeyParam {
			queryKeys = append(queryKeys, key)
		}
	}
	sort.Strings(queryKeys)

	for _, key := range queryKeys {
		values := queryParams[key]
		sort.Strings(values)
		for _, value := range values {
			parts = append(parts, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// 添加时间戳
	parts = append(parts, timestamp)

	// 添加 nonce
	if nonce != "" {
		parts = append(parts, nonce)
	}

	// 添加 appKey
	if appKey != "" {
		parts = append(parts, appKey)
	}

	// 添加请求体（如果需要）
	if config.ValidateBody && c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err == nil && len(bodyBytes) > 0 {
			parts = append(parts, string(bodyBytes))
			// 重新设置 Body，以便后续处理
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	return strings.Join(parts, "&")
}

// calculateSignature 计算签名
func calculateSignature(data, secretKey, algorithm string) string {
	switch algorithm {
	case "hmac-sha256":
		h := hmac.New(sha256.New, []byte(secretKey))
		h.Write([]byte(data))
		return hex.EncodeToString(h.Sum(nil))
	case "md5":
		h := md5.New()
		h.Write([]byte(data + secretKey))
		return hex.EncodeToString(h.Sum(nil))
	default:
		h := hmac.New(sha256.New, []byte(secretKey))
		h.Write([]byte(data))
		return hex.EncodeToString(h.Sum(nil))
	}
}

// cleanupNonce 清理过期的 nonce
func cleanupNonce(expiry time.Duration) {
	now := time.Now()
	for key, t := range nonceStore {
		if now.Sub(t) > expiry {
			delete(nonceStore, key)
		}
	}
}

// GenerateSignature 生成签名（供客户端使用）
func GenerateSignature(method, path string, params map[string]string, body string, secretKey string) (string, string, string) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := generateNonce()

	var parts []string
	parts = append(parts, method)
	parts = append(parts, path)

	// 排序参数
	var keys []string
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, params[key]))
	}

	parts = append(parts, timestamp)
	parts = append(parts, nonce)

	if body != "" {
		parts = append(parts, body)
	}

	signString := strings.Join(parts, "&")
	signature := calculateSignature(signString, secretKey, "hmac-sha256")

	return signature, timestamp, nonce
}

// generateNonce 生成随机 nonce
func generateNonce() string {
	h := md5.New()
	h.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// SimpleSignature 简单签名验证（仅验证 AppKey + Secret）
func SimpleSignature(appKeys map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		appKey := c.GetHeader("X-App-Key")
		signature := c.GetHeader("X-Signature")
		timestamp := c.GetHeader("X-Timestamp")

		if appKey == "" || signature == "" || timestamp == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "缺少认证参数",
			})
			c.Abort()
			return
		}

		secretKey, exists := appKeys[appKey]
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的 AppKey",
			})
			c.Abort()
			return
		}

		// 验证时间戳
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil || time.Now().Unix()-ts > 300 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求已过期",
			})
			c.Abort()
			return
		}

		// 简单签名：md5(appKey + timestamp + secretKey)
		expectedSign := calculateSignature(appKey+timestamp, secretKey, "md5")

		if signature != expectedSign {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "签名验证失败",
			})
			c.Abort()
			return
		}

		c.Set("app_key", appKey)
		c.Next()
	}
}
