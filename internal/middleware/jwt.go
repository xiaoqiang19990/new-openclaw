package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig JWT 配置
type JWTConfig struct {
	SecretKey     string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

// DefaultJWTConfig 默认 JWT 配置
var DefaultJWTConfig = JWTConfig{
	SecretKey:     "your-secret-key-change-in-production",
	TokenExpiry:   time.Hour * 24,
	RefreshExpiry: time.Hour * 24 * 7,
	Issuer:        "new-openclaw",
}

// Claims 自定义 JWT Claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTAuth JWT 认证中间件
func JWTAuth() gin.HandlerFunc {
	return JWTAuthWithConfig(DefaultJWTConfig)
}

// JWTAuthWithConfig 带配置的 JWT 认证中间件
func JWTAuthWithConfig(config JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少认证令牌",
			})
			c.Abort()
			return
		}

		// 检查 Bearer 前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "认证令牌格式错误",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析 Token
		claims, err := ParseToken(tokenString, config.SecretKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证令牌: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// GenerateToken 生成 JWT Token
func GenerateToken(userID, username, role string, config JWTConfig) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(config.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.SecretKey))
}

// GenerateRefreshToken 生成刷新 Token
func GenerateRefreshToken(userID string, config JWTConfig) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(config.RefreshExpiry)),
		IssuedAt:  jwt.NewNumericDate(now),
		Subject:   userID,
		Issuer:    config.Issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.SecretKey))
}

// ParseToken 解析 JWT Token
func ParseToken(tokenString, secretKey string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的令牌")
}

// OptionalJWTAuth 可选的 JWT 认证（不强制要求）
func OptionalJWTAuth() gin.HandlerFunc {
	return OptionalJWTAuthWithConfig(DefaultJWTConfig)
}

// OptionalJWTAuthWithConfig 带配置的可选 JWT 认证
func OptionalJWTAuthWithConfig(config JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := ParseToken(tokenString, config.SecretKey)
		if err == nil {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
			c.Set("claims", claims)
		}

		c.Next()
	}
}

// RequireRole 角色验证中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "未授权访问",
			})
			c.Abort()
			return
		}

		userRole := role.(string)
		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足",
		})
		c.Abort()
	}
}
