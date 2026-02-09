package database

import (
	"fmt"
	"log"
	"time"

	"new-openclaw/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var MySQL *gorm.DB

// InitMySQL 初始化 MySQL 连接
func InitMySQL(cfg *config.MySQLConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	var err error
	MySQL, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("连接 MySQL 失败: %w", err)
	}

	// 获取底层 sql.DB 设置连接池
	sqlDB, err := MySQL.DB()
	if err != nil {
		return fmt.Errorf("获取 MySQL 连接池失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期

	log.Println("✅ MySQL 连接成功")
	return nil
}

// CloseMySQL 关闭 MySQL 连接
func CloseMySQL() error {
	if MySQL != nil {
		sqlDB, err := MySQL.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetMySQL 获取 MySQL 实例
func GetMySQL() *gorm.DB {
	return MySQL
}
