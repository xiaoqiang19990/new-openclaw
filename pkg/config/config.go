package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig
	MySQL    MySQLConfig
	Redis    RedisConfig
	MongoDB  MongoDBConfig
	Security SecurityConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string
	Mode string
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	// JWT 配置
	JWTSecretKey     string
	JWTExpiry        time.Duration
	JWTRefreshExpiry time.Duration
	JWTIssuer        string

	// 频率限制配置
	RateLimitWindow      time.Duration
	RateLimitMaxRequests int

	// API 签名配置
	APISignatureKey    string
	APISignatureExpiry time.Duration

	// IP 过滤配置
	IPWhitelistMode bool
	IPWhitelist     []string
	IPBlacklist     []string

	// 审计配置
	AuditEnabled  bool
	AuditOutput   string
	AuditFilePath string
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// MongoDBConfig MongoDB 配置
type MongoDBConfig struct {
	URI      string
	Database string
}

// LoadConfig 加载配置（从环境变量）
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "localhost"),
			Port:     getEnv("MYSQL_PORT", "3306"),
			User:     getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", ""),
			DBName:   getEnv("MYSQL_DATABASE", "new_openclaw"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGO_DATABASE", "new_openclaw"),
		},
		Security: SecurityConfig{
			// JWT 配置
			JWTSecretKey:     getEnv("JWT_SECRET_KEY", "your-secret-key-change-in-production"),
			JWTExpiry:        getDurationEnv("JWT_EXPIRY", time.Hour*24),
			JWTRefreshExpiry: getDurationEnv("JWT_REFRESH_EXPIRY", time.Hour*24*7),
			JWTIssuer:        getEnv("JWT_ISSUER", "new-openclaw"),

			// 频率限制配置
			RateLimitWindow:      getDurationEnv("RATE_LIMIT_WINDOW", time.Minute),
			RateLimitMaxRequests: getIntEnv("RATE_LIMIT_MAX_REQUESTS", 60),

			// API 签名配置
			APISignatureKey:    getEnv("API_SIGNATURE_KEY", "your-api-secret-key"),
			APISignatureExpiry: getDurationEnv("API_SIGNATURE_EXPIRY", time.Minute*5),

			// IP 过滤配置
			IPWhitelistMode: getBoolEnv("IP_WHITELIST_MODE", false),
			IPWhitelist:     getSliceEnv("IP_WHITELIST", []string{}),
			IPBlacklist:     getSliceEnv("IP_BLACKLIST", []string{}),

			// 审计配置
			AuditEnabled:  getBoolEnv("AUDIT_ENABLED", true),
			AuditOutput:   getEnv("AUDIT_OUTPUT", "both"),
			AuditFilePath: getEnv("AUDIT_FILE_PATH", "logs/audit.log"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
