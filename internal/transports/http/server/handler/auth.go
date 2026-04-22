package handler

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandlerOpts struct {
	APIKey string
}

type AuthHandler struct {
	apiKey string
}

func NewAuthHandler(opts *AuthHandlerOpts) *AuthHandler {
	return &AuthHandler{apiKey: opts.APIKey}
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

	fmt.Println("KEY", req, h.apiKey)

	if subtle.ConstantTimeCompare([]byte(req.Key), []byte(h.apiKey)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": req.Key})
}
