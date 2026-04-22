package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sagarmaheshwary/reqlog-ui/internal/logger"
	"github.com/sagarmaheshwary/reqlog-ui/internal/service"
)

type ReqlogHandlerOpts struct {
	ReqlogService service.ReqlogService
	Logger        logger.Logger
}

type ReqlogHandler struct {
	reqlogService service.ReqlogService
	logger        logger.Logger
}

func NewReqlogHandler(opts *ReqlogHandlerOpts) *ReqlogHandler {
	return &ReqlogHandler{reqlogService: opts.ReqlogService, logger: opts.Logger}
}

func (h *ReqlogHandler) Logs(c *gin.Context) {
	params := parseParams(c)

	lines, err := h.reqlogService.Run(c.Request.Context(), params)
	if err != nil {
		h.logger.Error("reqlog run failed", logger.Field{Key: "error", Value: err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"lines": lines})
}

func parseParams(c *gin.Context) service.ReqlogParams {
	limitStr := c.DefaultQuery("limit", "0")
	limit, _ := strconv.Atoi(limitStr)

	recursive := true
	if r := c.Query("recursive"); r == "false" || r == "0" {
		recursive = false
	}

	return service.ReqlogParams{
		SearchValue: c.Query("q"),
		Dir:         c.DefaultQuery("dir", "./logs"),
		IgnoreCase:  c.Query("ignore_case") == "true" || c.Query("ignore_case") == "1",
		Limit:       limit,
		JSON:        c.Query("json") == "true" || c.Query("json") == "1",
		Key:         c.Query("key"),
		Since:       c.Query("since"),
		Recursive:   recursive,
		Service:     c.Query("service"),
		Source:      c.DefaultQuery("source", "file"),
	}
}
