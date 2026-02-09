package database

import (
	"log"

	"new-openclaw/internal/model"
	"new-openclaw/pkg/config"
)

// InitAll 初始化所有数据库连接
func InitAll(cfg *config.Config) error {
	// 初始化 MySQL（可选，连接失败只打印警告）
	if err := InitMySQL(&cfg.MySQL); err != nil {
		log.Printf("⚠️  MySQL 初始化失败（可选）: %v", err)
	} else {
		// 自动迁移数据库表
		if err := AutoMigrate(); err != nil {
			log.Printf("⚠️  数据库迁移失败: %v", err)
		}
	}

	// 初始化 Redis（可选，连接失败只打印警告）
	if err := InitRedis(&cfg.Redis); err != nil {
		log.Printf("⚠️  Redis 初始化失败（可选）: %v", err)
	}

	// 初始化 MongoDB（可选，连接失败只打印警告）
	if err := InitMongoDB(&cfg.MongoDB); err != nil {
		log.Printf("⚠️  MongoDB 初始化失败（可选）: %v", err)
	}

	return nil
}

// CloseAll 关闭所有数据库连接
func CloseAll() {
	if err := CloseMySQL(); err != nil {
		log.Printf("关闭 MySQL 失败: %v", err)
	}
	if err := CloseRedis(); err != nil {
		log.Printf("关闭 Redis 失败: %v", err)
	}
	if err := CloseMongoDB(); err != nil {
		log.Printf("关闭 MongoDB 失败: %v", err)
	}
	log.Println("✅ 所有数据库连接已关闭")
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate() error {
	if MySQL == nil {
		return nil
	}

	log.Println("🔄 开始数据库迁移...")

	// 迁移所有模型
	err := MySQL.AutoMigrate(
		&model.Admin{},
	)

	if err != nil {
		return err
	}

	log.Println("✅ 数据库迁移完成")
	return nil
}
