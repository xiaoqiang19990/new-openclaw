# New OpenClaw

一个基于 Gin 框架的 Golang HTTP 服务。

## 项目结构

```
new-openclaw/
├── cmd/
│   └── server/
│       └── main.go          # 程序入口
├── internal/
│   ├── handler/
│   │   ├── routes.go        # 路由注册
│   │   ├── health.go        # 健康检查接口
│   │   └── user.go          # 用户 CRUD 接口
│   └── middleware/
│       ├── logger.go        # 日志中间件
│       └── cors.go          # 跨域中间件
├── configs/                  # 配置文件目录
├── go.mod
├── Makefile
└── README.md
```

## 快速开始

### 安装依赖

```bash
go mod tidy
```

### 运行服务

```bash
# 方式1: 直接运行
go run cmd/server/main.go

# 方式2: 使用 Make
make run

# 方式3: 编译后运行
make build
./bin/server
```

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| PORT | 服务端口 | 8080 |
| GIN_MODE | 运行模式 (debug/release) | debug |

## API 接口

### 健康检查

```bash
# Ping
curl http://localhost:8080/ping

# 健康检查
curl http://localhost:8080/health
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

## 技术栈

- Go 1.21+
- Gin Web Framework
- 内存存储（可扩展为 MySQL/PostgreSQL）

## License

MIT
