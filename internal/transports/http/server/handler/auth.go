package handler

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sagarmaheshwary/reqlog-ui/internal/tokenstore"
)

type AuthHandlerOpts struct {
	APIKey     string
	TokenStore *tokenstore.Store
}

type AuthHandler struct {
	apiKey     string
	tokenStore *tokenstore.Store
}

func NewAuthHandler(opts *AuthHandlerOpts) *AuthHandler {
	return &AuthHandler{apiKey: opts.APIKey, tokenStore: opts.TokenStore}
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

	if subtle.ConstantTimeCompare([]byte(req.Key), []byte(h.apiKey)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": req.Key})
}

func (h *AuthHandler) StreamToken(c *gin.Context) {
	token, err := h.tokenStore.Issue()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
