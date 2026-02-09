package handler

import (
	"net/http"
	"strconv"

	"new-openclaw/internal/database"
	"new-openclaw/internal/model"

	"github.com/gin-gonic/gin"
)

// ListAdmins 获取管理员列表
// @Summary 获取管理员列表
// @Tags Admin
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} map[string]interface{}
// @Router /admin/admins [get]
func ListAdmins(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	db := database.GetMySQL()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库未连接",
		})
		return
	}

	var admins []model.Admin
	var total int64

	db.Model(&model.Admin{}).Count(&total)
	db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&admins)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"list":      admins,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// CreateAdmin 创建管理员
// @Summary 创建管理员
// @Tags Admin
// @Accept json
// @Produce json
// @Param body body map[string]interface{} true "管理员信息"
// @Success 200 {object} map[string]interface{}
// @Router /admin/admins [post]
func CreateAdmin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=50"`
		Password string `json:"password" binding:"required,min=6,max=50"`
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	db := database.GetMySQL()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库未连接",
		})
		return
	}

	// 检查用户名是否已存在
	var count int64
	db.Model(&model.Admin{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户名已存在",
		})
		return
	}

	admin := model.Admin{
		Username: req.Username,
		Nickname: req.Nickname,
		Email:    req.Email,
		Role:     req.Role,
		Status:   1,
	}

	if admin.Role == "" {
		admin.Role = "admin"
	}

	if err := admin.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "密码加密失败",
		})
		return
	}

	if err := db.Create(&admin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    admin,
	})
}

// UpdateAdmin 更新管理员
// @Summary 更新管理员
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "管理员ID"
// @Param body body map[string]interface{} true "管理员信息"
// @Success 200 {object} map[string]interface{}
// @Router /admin/admins/{id} [put]
func UpdateAdmin(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	var req struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Role     string `json:"role"`
		Status   *int   `json:"status"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	db := database.GetMySQL()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库未连接",
		})
		return
	}

	var admin model.Admin
	if err := db.First(&admin, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "管理员不存在",
		})
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Password != "" {
		if err := admin.SetPassword(req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "密码加密失败",
			})
			return
		}
		updates["password"] = admin.Password
	}

	if err := db.Model(&admin).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    admin,
	})
}

// DeleteAdmin 删除管理员
// @Summary 删除管理员
// @Tags Admin
// @Produce json
// @Param id path int true "管理员ID"
// @Success 200 {object} map[string]interface{}
// @Router /admin/admins/{id} [delete]
func DeleteAdmin(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID",
		})
		return
	}

	db := database.GetMySQL()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库未连接",
		})
		return
	}

	result := db.Delete(&model.Admin{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除失败: " + result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "管理员不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}
