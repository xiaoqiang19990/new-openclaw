package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Admin 管理员用户模型
type Admin struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Username  string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	Nickname  string         `gorm:"type:varchar(100)" json:"nickname"`
	Email     string         `gorm:"type:varchar(100);index" json:"email"`
	Avatar    string         `gorm:"type:varchar(255)" json:"avatar"`
	Role      string         `gorm:"type:varchar(20);default:admin" json:"role"` // super_admin, admin, editor
	Status    int            `gorm:"type:tinyint;default:1" json:"status"`       // 1: 启用, 0: 禁用
	LastLogin *time.Time     `json:"last_login"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Admin) TableName() string {
	return "admins"
}

// SetPassword 设置密码（加密）
func (a *Admin) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (a *Admin) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	return err == nil
}

// AdminLoginRequest 登录请求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
}

// AdminLoginResponse 登录响应
type AdminLoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	Admin     *Admin `json:"admin"`
}
