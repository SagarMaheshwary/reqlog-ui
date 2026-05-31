package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sagarmaheshwary/reqlog-ui/internal/logger"
	"github.com/sagarmaheshwary/reqlog-ui/internal/service"
)

type OptionsHandler struct {
	logger         logger.Logger
	optionsService *service.OptionsService
}

type OptionsHandlerOpts struct {
	Logger         logger.Logger
	OptionsService *service.OptionsService
}

func NewOptionsHandler(opts *OptionsHandlerOpts) *OptionsHandler {
	return &OptionsHandler{
		logger:         opts.Logger,
		optionsService: opts.OptionsService,
	}
}

func (h *OptionsHandler) ListDirectories(c *gin.Context) {
	dirs, err := h.optionsService.ListDirectories()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list directories"})
		return
	}

	c.JSON(200, gin.H{"directories": dirs})
}

func (h *OptionsHandler) ListFiles(c *gin.Context) {
	directory := c.Query("directory")
	files, err := h.optionsService.ListFiles(directory)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list files"})
		return
	}
	c.JSON(200, gin.H{"files": files})
}

func (h *OptionsHandler) ListContainers(c *gin.Context) {
	containers, err := h.optionsService.ListContainers()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list docker containers"})
		return
	}
	c.JSON(200, gin.H{"containers": containers})
}
