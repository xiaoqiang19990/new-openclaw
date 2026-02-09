package middleware

import (
	"net/http"
	"strings"

	"new-openclaw/pkg/jwt"

	"github.com/gin-gonic/gin"
)

const (
	// AdminContextKey 管理员信息在Context中的key
	AdminContextKey = "admin_claims"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "请先登录",
			})
			c.Abort()
			return
		}

		// 检查Token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token格式错误",
			})
			c.Abort()
			return
		}

		// 解析Token
		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			message := "Token无效"
			if err == jwt.ErrTokenExpired {
				message = "Token已过期，请重新登录"
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": message,
			})
			c.Abort()
			return
		}

		// 将管理员信息存入Context
		c.Set(AdminContextKey, claims)
		c.Next()
	}
}

// RequireRole 角色权限中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get(AdminContextKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "请先登录",
			})
			c.Abort()
			return
		}

		adminClaims := claims.(*jwt.Claims)
		
		// 检查角色权限
		hasRole := false
		for _, role := range roles {
			if adminClaims.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentAdmin 从Context获取当前管理员信息
func GetCurrentAdmin(c *gin.Context) *jwt.Claims {
	claims, exists := c.Get(AdminContextKey)
	if !exists {
		return nil
	}
	return claims.(*jwt.Claims)
}
