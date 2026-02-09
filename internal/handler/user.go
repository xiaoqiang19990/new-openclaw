package handler

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

// User 用户结构体
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age"`
}

// 模拟数据库（内存存储）
var (
	users  = make(map[int]*User)
	nextID = 1
	mu     sync.RWMutex
)

// GetUsers 获取所有用户
func GetUsers(c *gin.Context) {
	mu.RLock()
	defer mu.RUnlock()

	userList := make([]*User, 0, len(users))
	for _, u := range users {
		userList = append(userList, u)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    userList,
	})
}

// GetUserByID 根据 ID 获取用户
func GetUserByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户 ID",
		})
		return
	}

	mu.RLock()
	user, exists := users[id]
	mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    user,
	})
}

// CreateUser 创建用户
func CreateUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	mu.Lock()
	user.ID = nextID
	nextID++
	users[user.ID] = &user
	mu.Unlock()

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "创建成功",
		"data":    user,
	})
}

// UpdateUser 更新用户
func UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户 ID",
		})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := users[id]; !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	user.ID = id
	users[id] = &user

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
		"data":    user,
	})
}

// DeleteUser 删除用户
func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户 ID",
		})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := users[id]; !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	delete(users, id)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}
