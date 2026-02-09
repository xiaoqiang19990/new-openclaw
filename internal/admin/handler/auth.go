package handler

import (
	"net/http"
	"time"

	"new-openclaw/internal/database"
	"new-openclaw/internal/model"
	"new-openclaw/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// Login 管理员登录
// @Summary 管理员登录
// @Tags Admin
// @Accept json
// @Produce json
// @Param body body model.AdminLoginRequest true "登录信息"
// @Success 200 {object} model.AdminLoginResponse
// @Router /admin/login [post]
func Login(c *gin.Context) {
	var req model.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 查询管理员
	var admin model.Admin
	db := database.GetMySQL()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库未连接",
		})
		return
	}

	result := db.Where("username = ?", req.Username).First(&admin)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户名或密码错误",
		})
		return
	}

	// 检查状态
	if admin.Status != 1 {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "账号已被禁用",
		})
		return
	}

	// 验证密码
	if !admin.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户名或密码错误",
		})
		return
	}

	// 生成Token
	token, expiresAt, err := jwt.GenerateToken(admin.ID, admin.Username, admin.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成Token失败",
		})
		return
	}

	// 更新最后登录时间
	now := time.Now()
	db.Model(&admin).Update("last_login", now)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "登录成功",
		"data": model.AdminLoginResponse{
			Token:     token,
			ExpiresAt: expiresAt,
			Admin:     &admin,
		},
	})
}

// Logout 管理员登出
// @Summary 管理员登出
// @Tags Admin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /admin/logout [post]
func Logout(c *gin.Context) {
	// JWT是无状态的，客户端删除Token即可
	// 如需实现Token黑名单，可以将Token存入Redis
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "登出成功",
	})
}

// GetProfile 获取当前管理员信息
// @Summary 获取当前管理员信息
// @Tags Admin
// @Produce json
// @Success 200 {object} model.Admin
// @Router /admin/profile [get]
func GetProfile(c *gin.Context) {
	claims, exists := c.Get("admin_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "请先登录",
		})
		return
	}

	adminClaims := claims.(*jwt.Claims)

	var admin model.Admin
	db := database.GetMySQL()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库未连接",
		})
		return
	}

	result := db.First(&admin, adminClaims.AdminID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "管理员不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    admin,
	})
}

// RefreshToken 刷新Token
// @Summary 刷新Token
// @Tags Admin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /admin/refresh-token [post]
func RefreshToken(c *gin.Context) {
	claims, exists := c.Get("admin_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "请先登录",
		})
		return
	}

	adminClaims := claims.(*jwt.Claims)

	// 生成新Token
	token, expiresAt, err := jwt.GenerateToken(adminClaims.AdminID, adminClaims.Username, adminClaims.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "刷新Token失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"token":      token,
			"expires_at": expiresAt,
		},
	})
}
