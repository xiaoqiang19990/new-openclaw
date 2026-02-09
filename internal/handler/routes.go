package handler

import (
	"new-openclaw/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine) {
	// 健康检查（无需认证）
	r.GET("/ping", Ping)
	r.GET("/health", HealthCheck)

	// API v1 分组
	v1 := r.Group("/api/v1")
	{
		// 公开接口（无需认证，但有频率限制）
		public := v1.Group("/public")
		{
			public.POST("/login", Login)
			public.POST("/register", Register)
			public.POST("/refresh-token", RefreshToken)
		}

		// 需要 JWT 认证的接口
		auth := v1.Group("/")
		auth.Use(middleware.JWTAuth())
		{
			// 用户相关
			auth.GET("/users", GetUsers)
			auth.GET("/users/:id", GetUserByID)
			auth.POST("/users", CreateUser)
			auth.PUT("/users/:id", UpdateUser)
			auth.DELETE("/users/:id", DeleteUser)

			// 用户信息
			auth.GET("/profile", GetProfile)
			auth.PUT("/profile", UpdateProfile)
		}

		// 需要管理员权限的接口
		admin := v1.Group("/admin")
		admin.Use(middleware.JWTAuth())
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/users", GetAllUsers)
			admin.DELETE("/users/:id", AdminDeleteUser)
			admin.POST("/ip/blacklist", AddIPBlacklist)
			admin.DELETE("/ip/blacklist", RemoveIPBlacklist)
		}

		// 需要 API 签名验证的接口（用于第三方调用）
		signed := v1.Group("/signed")
		signed.Use(middleware.APISignature())
		{
			signed.POST("/webhook", HandleWebhook)
			signed.POST("/callback", HandleCallback)
		}
	}
}

// Login 用户登录
func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// TODO: 验证用户名密码
	// 这里仅作示例，实际应查询数据库验证
	if req.Username == "admin" && req.Password == "admin123" {
		token, err := middleware.GenerateToken("1", req.Username, "admin", middleware.DefaultJWTConfig)
		if err != nil {
			c.JSON(500, gin.H{
				"code":    500,
				"message": "生成令牌失败",
			})
			return
		}

		refreshToken, _ := middleware.GenerateRefreshToken("1", middleware.DefaultJWTConfig)

		c.JSON(200, gin.H{
			"code":    200,
			"message": "登录成功",
			"data": gin.H{
				"token":         token,
				"refresh_token": refreshToken,
				"expires_in":    86400,
			},
		})
		return
	}

	c.JSON(401, gin.H{
		"code":    401,
		"message": "用户名或密码错误",
	})
}

// Register 用户注册
func Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Email    string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// TODO: 实际注册逻辑
	c.JSON(200, gin.H{
		"code":    200,
		"message": "注册成功",
	})
}

// RefreshToken 刷新令牌
func RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// TODO: 验证 refresh token 并生成新 token
	c.JSON(200, gin.H{
		"code":    200,
		"message": "刷新成功",
	})
}

// GetProfile 获取当前用户信息
func GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("role")

	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"user_id":  userID,
			"username": username,
			"role":     role,
		},
	})
}

// UpdateProfile 更新当前用户信息
func UpdateProfile(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    200,
		"message": "更新成功",
	})
}

// GetAllUsers 管理员获取所有用户
func GetAllUsers(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"data":    []gin.H{},
	})
}

// AdminDeleteUser 管理员删除用户
func AdminDeleteUser(c *gin.Context) {
	id := c.Param("id")
	c.JSON(200, gin.H{
		"code":    200,
		"message": "用户 " + id + " 已删除",
	})
}

// AddIPBlacklist 添加 IP 黑名单
func AddIPBlacklist(c *gin.Context) {
	var req struct {
		IP string `json:"ip" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// TODO: 添加到黑名单
	c.JSON(200, gin.H{
		"code":    200,
		"message": "IP " + req.IP + " 已添加到黑名单",
	})
}

// RemoveIPBlacklist 移除 IP 黑名单
func RemoveIPBlacklist(c *gin.Context) {
	var req struct {
		IP string `json:"ip" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// TODO: 从黑名单移除
	c.JSON(200, gin.H{
		"code":    200,
		"message": "IP " + req.IP + " 已从黑名单移除",
	})
}

// HandleWebhook 处理 Webhook
func HandleWebhook(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    200,
		"message": "Webhook 处理成功",
	})
}

// HandleCallback 处理回调
func HandleCallback(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    200,
		"message": "Callback 处理成功",
	})
}
