package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditConfig 审计配置
type AuditConfig struct {
	// 是否启用
	Enabled bool
	// 日志输出方式：console, file, both
	Output string
	// 日志文件路径
	FilePath string
	// 是否记录请求体
	LogRequestBody bool
	// 是否记录响应体
	LogResponseBody bool
	// 请求体最大记录长度
	MaxRequestBodySize int
	// 响应体最大记录长度
	MaxResponseBodySize int
	// 敏感字段（会被脱敏）
	SensitiveFields []string
	// 排除的路径
	ExcludePaths []string
	// 自定义日志处理函数
	CustomHandler func(log *AuditLog)
	// 异步写入
	Async bool
	// 异步写入缓冲区大小
	BufferSize int
}

// DefaultAuditConfig 默认审计配置
var DefaultAuditConfig = AuditConfig{
	Enabled:             true,
	Output:              "both",
	FilePath:            "logs/audit.log",
	LogRequestBody:      true,
	LogResponseBody:     true,
	MaxRequestBodySize:  4096,
	MaxResponseBodySize: 4096,
	SensitiveFields:     []string{"password", "token", "secret", "key", "authorization"},
	ExcludePaths:        []string{"/ping", "/health", "/metrics"},
	Async:               true,
	BufferSize:          1000,
}

// AuditLog 审计日志结构
type AuditLog struct {
	// 请求 ID
	RequestID string `json:"request_id"`
	// 时间戳
	Timestamp time.Time `json:"timestamp"`
	// 客户端 IP
	ClientIP string `json:"client_ip"`
	// 用户 ID（如果已认证）
	UserID string `json:"user_id,omitempty"`
	// 用户名（如果已认证）
	Username string `json:"username,omitempty"`
	// 请求方法
	Method string `json:"method"`
	// 请求路径
	Path string `json:"path"`
	// 查询参数
	Query string `json:"query,omitempty"`
	// 请求头
	Headers map[string]string `json:"headers,omitempty"`
	// 请求体
	RequestBody string `json:"request_body,omitempty"`
	// 响应状态码
	StatusCode int `json:"status_code"`
	// 响应体
	ResponseBody string `json:"response_body,omitempty"`
	// 响应大小
	ResponseSize int `json:"response_size"`
	// 处理时间（毫秒）
	Latency int64 `json:"latency_ms"`
	// 错误信息
	Error string `json:"error,omitempty"`
	// User-Agent
	UserAgent string `json:"user_agent,omitempty"`
	// Referer
	Referer string `json:"referer,omitempty"`
	// 额外信息
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// responseWriter 自定义响应写入器（用于捕获响应体）
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// AuditLogger 审计日志记录器
type AuditLogger struct {
	config   AuditConfig
	file     *os.File
	logChan  chan *AuditLog
	mu       sync.Mutex
	wg       sync.WaitGroup
}

// NewAuditLogger 创建审计日志记录器
func NewAuditLogger(config AuditConfig) (*AuditLogger, error) {
	logger := &AuditLogger{
		config: config,
	}

	// 创建日志文件
	if config.Output == "file" || config.Output == "both" {
		dir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %v", err)
		}

		file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("打开日志文件失败: %v", err)
		}
		logger.file = file
	}

	// 异步模式
	if config.Async {
		logger.logChan = make(chan *AuditLog, config.BufferSize)
		logger.wg.Add(1)
		go logger.asyncWriter()
	}

	return logger, nil
}

// asyncWriter 异步写入协程
func (l *AuditLogger) asyncWriter() {
	defer l.wg.Done()

	for auditLog := range l.logChan {
		l.writeLog(auditLog)
	}
}

// writeLog 写入日志
func (l *AuditLogger) writeLog(auditLog *AuditLog) {
	l.mu.Lock()
	defer l.mu.Unlock()

	logJSON, err := json.Marshal(auditLog)
	if err != nil {
		log.Printf("审计日志序列化失败: %v", err)
		return
	}

	logLine := string(logJSON) + "\n"

	// 输出到控制台
	if l.config.Output == "console" || l.config.Output == "both" {
		log.Printf("[AUDIT] %s", logLine)
	}

	// 输出到文件
	if l.file != nil && (l.config.Output == "file" || l.config.Output == "both") {
		l.file.WriteString(logLine)
	}

	// 自定义处理
	if l.config.CustomHandler != nil {
		l.config.CustomHandler(auditLog)
	}
}

// Log 记录审计日志
func (l *AuditLogger) Log(auditLog *AuditLog) {
	if l.config.Async && l.logChan != nil {
		select {
		case l.logChan <- auditLog:
		default:
			// 缓冲区满，直接写入
			l.writeLog(auditLog)
		}
	} else {
		l.writeLog(auditLog)
	}
}

// Close 关闭日志记录器
func (l *AuditLogger) Close() {
	if l.logChan != nil {
		close(l.logChan)
		l.wg.Wait()
	}
	if l.file != nil {
		l.file.Close()
	}
}

// Audit 审计中间件
func Audit() gin.HandlerFunc {
	logger, err := NewAuditLogger(DefaultAuditConfig)
	if err != nil {
		log.Printf("创建审计日志记录器失败: %v", err)
		return func(c *gin.Context) { c.Next() }
	}

	return AuditWithLogger(logger)
}

// AuditWithConfig 带配置的审计中间件
func AuditWithConfig(config AuditConfig) gin.HandlerFunc {
	logger, err := NewAuditLogger(config)
	if err != nil {
		log.Printf("创建审计日志记录器失败: %v", err)
		return func(c *gin.Context) { c.Next() }
	}

	return AuditWithLogger(logger)
}

// AuditWithLogger 使用指定日志记录器的审计中间件
func AuditWithLogger(logger *AuditLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !logger.config.Enabled {
			c.Next()
			return
		}

		// 检查是否排除
		for _, path := range logger.config.ExcludePaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		startTime := time.Now()

		// 生成请求 ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// 读取请求体
		var requestBody string
		if logger.config.LogRequestBody && c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				if len(bodyBytes) > logger.config.MaxRequestBodySize {
					requestBody = string(bodyBytes[:logger.config.MaxRequestBodySize]) + "...(truncated)"
				} else {
					requestBody = string(bodyBytes)
				}
				// 脱敏处理
				requestBody = maskSensitiveData(requestBody, logger.config.SensitiveFields)
				// 重新设置 Body
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// 包装响应写入器
		var responseBody string
		if logger.config.LogResponseBody {
			rw := &responseWriter{
				ResponseWriter: c.Writer,
				body:           bytes.NewBuffer(nil),
			}
			c.Writer = rw

			c.Next()

			// 获取响应体
			if rw.body.Len() > logger.config.MaxResponseBodySize {
				responseBody = rw.body.String()[:logger.config.MaxResponseBodySize] + "...(truncated)"
			} else {
				responseBody = rw.body.String()
			}
			responseBody = maskSensitiveData(responseBody, logger.config.SensitiveFields)
		} else {
			c.Next()
		}

		// 构建审计日志
		auditLog := &AuditLog{
			RequestID:    requestID,
			Timestamp:    startTime,
			ClientIP:     c.ClientIP(),
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			Query:        c.Request.URL.RawQuery,
			RequestBody:  requestBody,
			StatusCode:   c.Writer.Status(),
			ResponseBody: responseBody,
			ResponseSize: c.Writer.Size(),
			Latency:      time.Since(startTime).Milliseconds(),
			UserAgent:    c.Request.UserAgent(),
			Referer:      c.Request.Referer(),
		}

		// 获取用户信息
		if userID, exists := c.Get("user_id"); exists {
			auditLog.UserID = userID.(string)
		}
		if username, exists := c.Get("username"); exists {
			auditLog.Username = username.(string)
		}

		// 获取错误信息
		if len(c.Errors) > 0 {
			auditLog.Error = c.Errors.String()
		}

		// 获取重要请求头
		auditLog.Headers = map[string]string{
			"Content-Type":  c.GetHeader("Content-Type"),
			"Authorization": maskString(c.GetHeader("Authorization")),
			"X-App-Key":     c.GetHeader("X-App-Key"),
		}

		// 记录日志
		logger.Log(auditLog)
	}
}

// generateRequestID 生成请求 ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond()%10000)
}

// maskSensitiveData 脱敏敏感数据
func maskSensitiveData(data string, sensitiveFields []string) string {
	for _, field := range sensitiveFields {
		// 简单的 JSON 字段脱敏
		data = maskJSONField(data, field)
	}
	return data
}

// maskJSONField 脱敏 JSON 字段
func maskJSONField(data, field string) string {
	// 这是一个简化的实现，实际使用可能需要更复杂的处理
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return data
	}

	maskMapField(result, field)

	masked, err := json.Marshal(result)
	if err != nil {
		return data
	}
	return string(masked)
}

// maskMapField 递归脱敏 map 字段
func maskMapField(data map[string]interface{}, field string) {
	for key, value := range data {
		if key == field {
			data[key] = "***MASKED***"
		} else if nested, ok := value.(map[string]interface{}); ok {
			maskMapField(nested, field)
		}
	}
}

// maskString 脱敏字符串
func maskString(s string) string {
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "***" + s[len(s)-4:]
}

// SecurityAudit 安全审计中间件（记录安全相关事件）
func SecurityAudit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录安全相关信息
		securityLog := map[string]interface{}{
			"timestamp":  time.Now(),
			"client_ip":  c.ClientIP(),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"user_agent": c.Request.UserAgent(),
		}

		// 检测可疑行为
		suspicious := false
		var reasons []string

		// 检查 SQL 注入特征
		query := c.Request.URL.RawQuery
		if containsSQLInjection(query) {
			suspicious = true
			reasons = append(reasons, "可能的 SQL 注入")
		}

		// 检查 XSS 特征
		if containsXSS(query) {
			suspicious = true
			reasons = append(reasons, "可能的 XSS 攻击")
		}

		// 检查路径遍历
		if containsPathTraversal(c.Request.URL.Path) {
			suspicious = true
			reasons = append(reasons, "可能的路径遍历")
		}

		if suspicious {
			securityLog["suspicious"] = true
			securityLog["reasons"] = reasons
			log.Printf("[SECURITY ALERT] %v", securityLog)
		}

		c.Next()
	}
}

// containsSQLInjection 检查是否包含 SQL 注入特征
func containsSQLInjection(s string) bool {
	patterns := []string{
		"'--", "' OR ", "' AND ", "UNION SELECT", "DROP TABLE",
		"INSERT INTO", "DELETE FROM", "UPDATE SET", "1=1", "1'='1",
	}
	for _, p := range patterns {
		if bytes.Contains(bytes.ToUpper([]byte(s)), []byte(p)) {
			return true
		}
	}
	return false
}

// containsXSS 检查是否包含 XSS 特征
func containsXSS(s string) bool {
	patterns := []string{
		"<script", "javascript:", "onerror=", "onload=", "onclick=",
		"<iframe", "<object", "<embed", "expression(",
	}
	for _, p := range patterns {
		if bytes.Contains(bytes.ToLower([]byte(s)), []byte(p)) {
			return true
		}
	}
	return false
}

// containsPathTraversal 检查是否包含路径遍历特征
func containsPathTraversal(s string) bool {
	patterns := []string{
		"../", "..\\", "%2e%2e", "%252e%252e",
	}
	for _, p := range patterns {
		if bytes.Contains(bytes.ToLower([]byte(s)), []byte(p)) {
			return true
		}
	}
	return false
}
