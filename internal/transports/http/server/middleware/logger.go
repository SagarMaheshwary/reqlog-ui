package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func ZerologMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		event := log.Info()
		if status >= 400 {
			event = log.Error()
		}

		event.
			Str("method", method).
			Str("path", path).
			Str("query", raw).
			Str("client_ip", clientIP).
			Int("status", status).
			Dur("latency", latency).
			Msg("incoming request")
	}
}
