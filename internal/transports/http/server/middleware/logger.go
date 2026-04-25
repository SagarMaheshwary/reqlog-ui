package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sagarmaheshwary/reqlog-ui/internal/logger"
)

func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		fields := []logger.Field{
			{Key: "method", Value: method},
			{Key: "path", Value: path},
			{Key: "query", Value: raw},
			{Key: "client_ip", Value: clientIP},
			{Key: "status", Value: status},
			{Key: "latency", Value: fmt.Sprintf("%.2fms", float64(latency.Microseconds())/1000)},
		}

		if status >= 400 {
			log.Error("http request completed", fields...)
			return
		}

		log.Info("http request completed", fields...)
	}
}
