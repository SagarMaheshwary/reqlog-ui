package handler

import (
	"crypto/subtle"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sagarmaheshwary/reqlog-ui/internal/config"
)

type AuthHandlerOpts struct {
	Config *config.HTTPServer
}

type AuthHandler struct {
	config *config.HTTPServer
}

func NewAuthHandler(opts *AuthHandlerOpts) *AuthHandler {
	return &AuthHandler{
		config: opts.Config,
	}
}

type tokenRequest struct {
	Key string `json:"key" binding:"required"`
}

func (h *AuthHandler) Token(c *gin.Context) {
	var req tokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	if subtle.ConstantTimeCompare([]byte(req.Key), []byte(h.config.APIKey)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid key"})
		return
	}

	expiry := 86400 // 24h

	claims := jwt.MapClaims{
		"sub": "reqlog-ui",
		"exp": time.Now().Add(time.Duration(expiry) * time.Second).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		log.Println(err, h.config.JWTSecret)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	c.SetCookie(
		config.AuthCookieName, // name
		tokenString,           // value
		expiry,                // max age
		"/",                   // path
		"",                    // domain
		h.config.HTTPS,        // secure (true in HTTPS)
		true,                  // httpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie(
		config.AuthCookieName,
		"",
		-1,
		"/",
		"",
		h.config.HTTPS,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func IsAuthenticated(c *gin.Context, jwtSecret []byte) bool {
	tokenString, err := c.Cookie(config.AuthCookieName)
	if err != nil {
		return false
	}

	_, err = jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return jwtSecret, nil
	})

	return err == nil
}
