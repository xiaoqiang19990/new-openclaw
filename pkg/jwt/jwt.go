package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired     = errors.New("token已过期")
	ErrTokenNotValidYet = errors.New("token尚未生效")
	ErrTokenMalformed   = errors.New("token格式错误")
	ErrTokenInvalid     = errors.New("无效的token")
)

// Config JWT配置
type Config struct {
	SecretKey     string
	ExpireHours   int
	Issuer        string
	TokenPrefix   string
}

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	SecretKey:   "openclaw-admin-secret-key-2024",
	ExpireHours: 24,
	Issuer:      "openclaw-admin",
	TokenPrefix: "Bearer ",
}

// Claims 自定义声明
type Claims struct {
	AdminID  uint   `json:"admin_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT Token
func GenerateToken(adminID uint, username, role string) (string, int64, error) {
	return GenerateTokenWithConfig(adminID, username, role, DefaultConfig)
}

// GenerateTokenWithConfig 使用自定义配置生成Token
func GenerateTokenWithConfig(adminID uint, username, role string, cfg *Config) (string, int64, error) {
	expiresAt := time.Now().Add(time.Duration(cfg.ExpireHours) * time.Hour)
	
	claims := &Claims{
		AdminID:  adminID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.SecretKey))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}

// ParseToken 解析JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	return ParseTokenWithConfig(tokenString, DefaultConfig)
}

// ParseTokenWithConfig 使用自定义配置解析Token
func ParseTokenWithConfig(tokenString string, cfg *Config) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotValidYet
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		}
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// RefreshToken 刷新Token
func RefreshToken(tokenString string) (string, int64, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", 0, err
	}
	return GenerateToken(claims.AdminID, claims.Username, claims.Role)
}
