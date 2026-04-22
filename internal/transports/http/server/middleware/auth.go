// internal/transports/http/server/middleware/auth.go
package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sagarmaheshwary/reqlog-ui/internal/tokenstore"
)

func APIKeyAuth(validKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearer(c.GetHeader("Authorization"))
		if token == "" {
			token = c.GetHeader("X-API-Key")
		}

		if subtle.ConstantTimeCompare([]byte(token), []byte(validKey)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		c.Next()
	}
}

func StreamTokenAuth(store *tokenstore.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" || !store.Consume(token) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.Next()
	}
}

func extractBearer(header string) string {
	const prefix = "Bearer "
	if strings.HasPrefix(header, prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return ""
}
