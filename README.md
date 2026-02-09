# New OpenClaw

一个基于 Gin 框架的 Golang HTTP 服务，集成 MySQL、Redis、MongoDB。

## 项目结构

```
new-openclaw/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── internal/
│   ├── database/
│   │   ├── init.go              # 数据库初始化
│   │   ├── mysql.go             # MySQL 连接
│   │   ├── redis.go             # Redis 连接
│   │   └── mongodb.go           # MongoDB 连接
│   ├── handler/
│   │   ├── routes.go            # 路由注册
│   │   ├── health.go            # 健康检查接口
│   │   └── user.go              # 用户 CRUD 接口
│   └── middleware/
│       ├── logger.go            # 日志中间件
│       └── cors.go              # 跨域中间件
├── pkg/
│   └── config/
│       └── config.go            # 配置管理
├── .env.example                  # 环境变量示例
├── go.mod
├── Makefile
└── README.md
```

## 技术栈

- **Web 框架**: Gin v1.9
- **MySQL**: GORM v1.25
- **Redis**: go-redis v8
- **MongoDB**: mongo-driver v1.13

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，填入你的数据库配置
```

### 3. 运行服务

```bash
# 方式1: 直接运行
go run cmd/server/main.go

# 方式2: 使用 Make
make run

# 方式3: 编译后运行
make build
./bin/server
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| PORT | 服务端口 | 8080 |
| GIN_MODE | 运行模式 | debug |
| MYSQL_HOST | MySQL 主机 | localhost |
| MYSQL_PORT | MySQL 端口 | 3306 |
| MYSQL_USER | MySQL 用户 | root |
| MYSQL_PASSWORD | MySQL 密码 | - |
| MYSQL_DATABASE | MySQL 数据库 | new_openclaw |
| REDIS_HOST | Redis 主机 | localhost |
| REDIS_PORT | Redis 端口 | 6379 |
| REDIS_PASSWORD | Redis 密码 | - |
| MONGO_URI | MongoDB URI | mongodb://localhost:27017 |
| MONGO_DATABASE | MongoDB 数据库 | new_openclaw |

## API 接口

### 健康检查

```bash
# Ping
curl http://localhost:8080/ping

# 健康检查（包含数据库状态）
curl http://localhost:8080/health
```

响应示例：
```json
{
  "status": "ok",
  "timestamp": "2026-02-09T12:00:00Z",
  "service": "new-openclaw",
  "version": "1.0.0",
  "mysql": "connected",
  "redis": "connected",
  "mongodb": "connected"
}
```

### 用户管理

```bash
# 获取所有用户
curl http://localhost:8080/api/v1/users

# 获取单个用户
curl http://localhost:8080/api/v1/users/1

# 创建用户
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name": "张三", "email": "zhangsan@example.com", "age": 25}'

# 更新用户
curl -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "张三", "email": "zhangsan@example.com", "age": 26}'

# 删除用户
curl -X DELETE http://localhost:8080/api/v1/users/1
```

## 数据库连接特性

### MySQL (GORM)
- 连接池: 最大100连接，最小10空闲
- 自动重连
- 日志记录

### Redis
- 连接池: 100连接，10最小空闲
- 超时配置: 连接5s，读写3s

### MongoDB
- 连接池: 最大100，最小10
- 自动重连
- 支持副本集

## License

MIT
