package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"new-openclaw/pkg/config"

	"github.com/go-redis/redis/v8"
)

var Redis *redis.Client

// InitRedis 初始化 Redis 连接
func InitRedis(cfg *config.RedisConfig) error {
	Redis = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     100,              // 连接池大小
		MinIdleConns: 10,               // 最小空闲连接数
		DialTimeout:  5 * time.Second,  // 连接超时
		ReadTimeout:  3 * time.Second,  // 读超时
		WriteTimeout: 3 * time.Second,  // 写超时
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("连接 Redis 失败: %w", err)
	}

	log.Println("✅ Redis 连接成功")
	return nil
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() error {
	if Redis != nil {
		return Redis.Close()
	}
	return nil
}

// GetRedis 获取 Redis 实例
func GetRedis() *redis.Client {
	return Redis
}
