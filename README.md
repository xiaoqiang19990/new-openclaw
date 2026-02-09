# New OpenClaw

一个基于 Gin 框架的 Golang HTTP 服务，集成 MySQL、Redis、MongoDB，并提供完整的 API 安全检测功能。

## 项目结构

```
new-openclaw/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── internal/
│   ├── admin/                   # 管理后台
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
│       ├── cors.go              # 跨域中间件
│       ├── jwt.go               # JWT Token 认证中间件
│       ├── ratelimit.go         # 请求频率限制中间件
│       ├── signature.go         # API 签名验证中间件
│       ├── ipfilter.go          # IP 白名单/黑名单中间件
│       ├── audit.go             # 请求日志审计中间件
│       └── security.go          # 安全中间件统一入口
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
- **JWT**: golang-jwt v5

## 🔐 API 安全功能

### 1. JWT Token 认证

支持 Bearer Token 认证，包含：
- Access Token 生成与验证
- Refresh Token 刷新机制
- 角色权限验证
- 可选认证模式

```go
// 使用示例
auth := v1.Group("/")
auth.Use(middleware.JWTAuth())

// 角色验证
admin.Use(middleware.RequireRole("admin"))
```

### 2. 请求频率限制 (Rate Limiting)

支持多种限流策略：
- 固定窗口限流
- 滑动窗口限流
- 基于 IP 限流
- 基于用户 ID 限流
- 基于端点限流

```go
// 全局限流：每分钟 60 次
r.Use(middleware.RateLimit())

// 自定义限流
r.Use(middleware.APIRateLimit(100, time.Minute))

// 滑动窗口限流
r.Use(middleware.SlidingWindowRateLimit(60, time.Minute))
```

### 3. API 签名验证

支持 HMAC-SHA256 和 MD5 签名算法：
- 时间戳验证（防止过期请求）
- Nonce 验证（防止重放攻击）
- 请求体签名
- 多 AppKey 支持

```bash
# 请求示例
curl -X POST http://localhost:8080/api/v1/signed/webhook \
  -H "X-App-Key: your-app-key" \
  -H "X-Timestamp: 1707480000" \
  -H "X-Nonce: abc123" \
  -H "X-Signature: calculated-signature" \
  -d '{"data": "test"}'
```

### 4. IP 白名单/黑名单

支持动态 IP 过滤：
- 白名单模式（只允许指定 IP）
- 黑名单模式（阻止指定 IP）
- CIDR 网段支持
- 私有 IP 自动放行
- 代理头信任配置
- 运行时动态添加/移除

```go
// 白名单模式
r.Use(middleware.IPWhitelist("192.168.1.0/24", "10.0.0.1"))

// 黑名单模式
r.Use(middleware.IPBlacklist("1.2.3.4", "5.6.7.0/24"))

// 动态管理
filter := middleware.NewDynamicIPFilter(config)
filter.AddBlacklist("1.2.3.4")
filter.RemoveBlacklist("1.2.3.4")
```

### 5. 请求日志审计

完整的请求审计功能：
- 请求/响应体记录
- 敏感数据脱敏
- 异步写入（高性能）
- 多输出方式（控制台/文件）
- 安全攻击检测（SQL注入、XSS、路径遍历）

```json
{
  "request_id": "1707480000-1234",
  "timestamp": "2026-02-09T12:00:00Z",
  "client_ip": "192.168.1.100",
  "user_id": "1",
  "method": "POST",
  "path": "/api/v1/users",
  "status_code": 200,
  "latency_ms": 15,
  "request_body": "{\"name\":\"test\",\"password\":\"***MASKED***\"}",
  "response_body": "{\"code\":200,\"message\":\"success\"}"
}
```

### 6. 安全响应头

自动添加安全响应头：
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Content-Security-Policy: default-src 'self'`
- `Strict-Transport-Security` (HSTS)

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，填入你的配置
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

### 基础配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| PORT | 服务端口 | 8080 |
| GIN_MODE | 运行模式 | debug |

### 数据库配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
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

### 安全配置

| 变量 | 说明 | 默认值 |
|------|------|--------|
| JWT_SECRET_KEY | JWT 密钥 | your-secret-key... |
| JWT_EXPIRY | Token 有效期 | 24h |
| JWT_REFRESH_EXPIRY | 刷新 Token 有效期 | 168h |
| JWT_ISSUER | Token 签发者 | new-openclaw |
| RATE_LIMIT_WINDOW | 限流时间窗口 | 1m |
| RATE_LIMIT_MAX_REQUESTS | 窗口内最大请求数 | 60 |
| API_SIGNATURE_KEY | API 签名密钥 | your-api-secret-key |
| API_SIGNATURE_EXPIRY | 签名有效期 | 5m |
| IP_WHITELIST_MODE | 白名单模式 | false |
| IP_WHITELIST | IP 白名单（逗号分隔） | - |
| IP_BLACKLIST | IP 黑名单（逗号分隔） | - |
| AUDIT_ENABLED | 启用审计日志 | true |
| AUDIT_OUTPUT | 审计输出方式 | both |
| AUDIT_FILE_PATH | 审计日志文件路径 | logs/audit.log |

## API 接口

### 公开接口

```bash
# Ping
curl http://localhost:8080/ping

# 健康检查
curl http://localhost:8080/health

# 用户登录
curl -X POST http://localhost:8080/api/v1/public/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'

# 用户注册
curl -X POST http://localhost:8080/api/v1/public/register \
  -H "Content-Type: application/json" \
  -d '{"username": "test", "password": "test123", "email": "test@example.com"}'
```

### 认证接口

```bash
# 获取用户信息（需要 JWT Token）
curl http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer <your-token>"

# 获取所有用户
curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <your-token>"
```

### 管理员接口

```bash
# 添加 IP 黑名单（需要管理员权限）
curl -X POST http://localhost:8080/api/v1/admin/ip/blacklist \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"ip": "1.2.3.4"}'
```

### 签名验证接口

```bash
# Webhook 回调（需要 API 签名）
curl -X POST http://localhost:8080/api/v1/signed/webhook \
  -H "X-App-Key: your-app-key" \
  -H "X-Timestamp: $(date +%s)" \
  -H "X-Nonce: $(openssl rand -hex 8)" \
  -H "X-Signature: <calculated-signature>" \
  -d '{"event": "test"}'
```

## 安全最佳实践

1. **生产环境必须修改默认密钥**
   - `JWT_SECRET_KEY`
   - `API_SIGNATURE_KEY`

2. **启用 HTTPS**
   - 配合 Nginx/Caddy 等反向代理

3. **配置合理的频率限制**
   - 根据业务需求调整 `RATE_LIMIT_MAX_REQUESTS`

4. **定期审查审计日志**
   - 关注安全告警日志
   - 监控异常访问模式

5. **IP 过滤策略**
   - 管理接口建议启用 IP 白名单
   - 及时更新黑名单

## License

MIT
