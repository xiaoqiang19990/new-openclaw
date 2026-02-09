package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"new-openclaw/pkg/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var MongoDB *mongo.Database
var MongoClient *mongo.Client

// InitMongoDB 初始化 MongoDB 连接
func InitMongoDB(cfg *config.MongoDBConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 设置客户端选项
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetConnectTimeout(10 * time.Second)

	// 连接 MongoDB
	var err error
	MongoClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("连接 MongoDB 失败: %w", err)
	}

	// 测试连接
	if err := MongoClient.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("MongoDB Ping 失败: %w", err)
	}

	// 获取数据库实例
	MongoDB = MongoClient.Database(cfg.Database)

	log.Println("✅ MongoDB 连接成功")
	return nil
}

// CloseMongoDB 关闭 MongoDB 连接
func CloseMongoDB() error {
	if MongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return MongoClient.Disconnect(ctx)
	}
	return nil
}

// GetMongoDB 获取 MongoDB 数据库实例
func GetMongoDB() *mongo.Database {
	return MongoDB
}

// GetMongoCollection 获取 MongoDB 集合
func GetMongoCollection(name string) *mongo.Collection {
	return MongoDB.Collection(name)
}
