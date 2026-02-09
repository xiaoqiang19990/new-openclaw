package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// 请求信息
		method := c.Request.Method
		uri := c.Request.RequestURI
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		// 打印日志
		log.Printf("[%s] %s %s %d %v",
			clientIP,
			method,
			uri,
			statusCode,
			latency,
		)
	}
}
