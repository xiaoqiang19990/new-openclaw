package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// IPFilterConfig IP 过滤配置
type IPFilterConfig struct {
	// 白名单模式（true: 只允许白名单，false: 只阻止黑名单）
	WhitelistMode bool
	// 白名单 IP 列表
	Whitelist []string
	// 黑名单 IP 列表
	Blacklist []string
	// 是否允许私有 IP
	AllowPrivate bool
	// 是否信任代理头
	TrustProxy bool
	// 代理头名称
	ProxyHeader string
	// 被阻止时的响应
	BlockHandler gin.HandlerFunc
}

// DefaultIPFilterConfig 默认 IP 过滤配置
var DefaultIPFilterConfig = IPFilterConfig{
	WhitelistMode: false,
	Whitelist:     []string{},
	Blacklist:     []string{},
	AllowPrivate:  true,
	TrustProxy:    true,
	ProxyHeader:   "X-Real-IP",
	BlockHandler: func(c *gin.Context) {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "IP 地址被禁止访问",
		})
		c.Abort()
	},
}

// IPFilter IP 过滤器
type IPFilter struct {
	config     IPFilterConfig
	whitelist  map[string]bool
	blacklist  map[string]bool
	whiteNets  []*net.IPNet
	blackNets  []*net.IPNet
	mu         sync.RWMutex
}

// NewIPFilter 创建 IP 过滤器
func NewIPFilter(config IPFilterConfig) *IPFilter {
	filter := &IPFilter{
		config:    config,
		whitelist: make(map[string]bool),
		blacklist: make(map[string]bool),
	}

	// 解析白名单
	for _, ip := range config.Whitelist {
		if strings.Contains(ip, "/") {
			_, ipNet, err := net.ParseCIDR(ip)
			if err == nil {
				filter.whiteNets = append(filter.whiteNets, ipNet)
			}
		} else {
			filter.whitelist[ip] = true
		}
	}

	// 解析黑名单
	for _, ip := range config.Blacklist {
		if strings.Contains(ip, "/") {
			_, ipNet, err := net.ParseCIDR(ip)
			if err == nil {
				filter.blackNets = append(filter.blackNets, ipNet)
			}
		} else {
			filter.blacklist[ip] = true
		}
	}

	return filter
}

// IsAllowed 检查 IP 是否允许
func (f *IPFilter) IsAllowed(ip string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 白名单模式
	if f.config.WhitelistMode {
		// 检查是否在白名单中
		if f.whitelist[ip] {
			return true
		}
		for _, ipNet := range f.whiteNets {
			if ipNet.Contains(parsedIP) {
				return true
			}
		}
		// 如果允许私有 IP，检查是否为私有 IP
		if f.config.AllowPrivate && isPrivateIP(parsedIP) {
			return true
		}
		return false
	}

	// 黑名单模式
	if f.blacklist[ip] {
		return false
	}
	for _, ipNet := range f.blackNets {
		if ipNet.Contains(parsedIP) {
			return false
		}
	}

	return true
}

// AddToWhitelist 添加到白名单
func (f *IPFilter) AddToWhitelist(ip string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if strings.Contains(ip, "/") {
		_, ipNet, err := net.ParseCIDR(ip)
		if err == nil {
			f.whiteNets = append(f.whiteNets, ipNet)
		}
	} else {
		f.whitelist[ip] = true
	}
}

// AddToBlacklist 添加到黑名单
func (f *IPFilter) AddToBlacklist(ip string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if strings.Contains(ip, "/") {
		_, ipNet, err := net.ParseCIDR(ip)
		if err == nil {
			f.blackNets = append(f.blackNets, ipNet)
		}
	} else {
		f.blacklist[ip] = true
	}
}

// RemoveFromWhitelist 从白名单移除
func (f *IPFilter) RemoveFromWhitelist(ip string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.whitelist, ip)
}

// RemoveFromBlacklist 从黑名单移除
func (f *IPFilter) RemoveFromBlacklist(ip string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.blacklist, ip)
}

// isPrivateIP 检查是否为私有 IP
func isPrivateIP(ip net.IP) bool {
	privateBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"::1/128",
		"fc00::/7",
	}

	for _, block := range privateBlocks {
		_, ipNet, _ := net.ParseCIDR(block)
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}

// getClientIP 获取客户端真实 IP
func getClientIP(c *gin.Context, config IPFilterConfig) string {
	if config.TrustProxy {
		// 尝试从代理头获取
		if ip := c.GetHeader(config.ProxyHeader); ip != "" {
			return strings.TrimSpace(strings.Split(ip, ",")[0])
		}
		if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
			return strings.TrimSpace(strings.Split(ip, ",")[0])
		}
	}
	return c.ClientIP()
}

// IPFilterMiddleware IP 过滤中间件
func IPFilterMiddleware() gin.HandlerFunc {
	return IPFilterWithConfig(DefaultIPFilterConfig)
}

// IPFilterWithConfig 带配置的 IP 过滤中间件
func IPFilterWithConfig(config IPFilterConfig) gin.HandlerFunc {
	filter := NewIPFilter(config)

	return func(c *gin.Context) {
		ip := getClientIP(c, config)

		if !filter.IsAllowed(ip) {
			config.BlockHandler(c)
			return
		}

		// 将 IP 存入上下文
		c.Set("client_ip", ip)
		c.Next()
	}
}

// IPWhitelist 白名单中间件
func IPWhitelist(ips ...string) gin.HandlerFunc {
	config := IPFilterConfig{
		WhitelistMode: true,
		Whitelist:     ips,
		AllowPrivate:  true,
		TrustProxy:    true,
		ProxyHeader:   "X-Real-IP",
		BlockHandler:  DefaultIPFilterConfig.BlockHandler,
	}

	return IPFilterWithConfig(config)
}

// IPBlacklist 黑名单中间件
func IPBlacklist(ips ...string) gin.HandlerFunc {
	config := IPFilterConfig{
		WhitelistMode: false,
		Blacklist:     ips,
		AllowPrivate:  true,
		TrustProxy:    true,
		ProxyHeader:   "X-Real-IP",
		BlockHandler:  DefaultIPFilterConfig.BlockHandler,
	}

	return IPFilterWithConfig(config)
}

// DynamicIPFilter 动态 IP 过滤器（支持运行时修改）
type DynamicIPFilter struct {
	filter *IPFilter
}

// NewDynamicIPFilter 创建动态 IP 过滤器
func NewDynamicIPFilter(config IPFilterConfig) *DynamicIPFilter {
	return &DynamicIPFilter{
		filter: NewIPFilter(config),
	}
}

// Middleware 返回中间件
func (d *DynamicIPFilter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c, d.filter.config)

		if !d.filter.IsAllowed(ip) {
			d.filter.config.BlockHandler(c)
			return
		}

		c.Set("client_ip", ip)
		c.Next()
	}
}

// AddWhitelist 添加白名单
func (d *DynamicIPFilter) AddWhitelist(ip string) {
	d.filter.AddToWhitelist(ip)
}

// AddBlacklist 添加黑名单
func (d *DynamicIPFilter) AddBlacklist(ip string) {
	d.filter.AddToBlacklist(ip)
}

// RemoveWhitelist 移除白名单
func (d *DynamicIPFilter) RemoveWhitelist(ip string) {
	d.filter.RemoveFromWhitelist(ip)
}

// RemoveBlacklist 移除黑名单
func (d *DynamicIPFilter) RemoveBlacklist(ip string) {
	d.filter.RemoveFromBlacklist(ip)
}

// CountryFilter 国家/地区过滤（需要 GeoIP 数据库支持）
// 这里提供接口，实际使用需要集成 GeoIP 库
type CountryFilter struct {
	AllowedCountries []string
	BlockedCountries []string
	// GeoIP 查询函数（需要外部实现）
	LookupFunc func(ip string) string
}

// CountryFilterMiddleware 国家过滤中间件
func CountryFilterMiddleware(cf *CountryFilter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cf.LookupFunc == nil {
			c.Next()
			return
		}

		ip := c.ClientIP()
		country := cf.LookupFunc(ip)

		// 检查是否在允许列表
		if len(cf.AllowedCountries) > 0 {
			allowed := false
			for _, ac := range cf.AllowedCountries {
				if ac == country {
					allowed = true
					break
				}
			}
			if !allowed {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "您所在的地区无法访问",
				})
				c.Abort()
				return
			}
		}

		// 检查是否在阻止列表
		for _, bc := range cf.BlockedCountries {
			if bc == country {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "您所在的地区无法访问",
				})
				c.Abort()
				return
			}
		}

		c.Set("country", country)
		c.Next()
	}
}
