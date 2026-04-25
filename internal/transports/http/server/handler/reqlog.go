package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sagarmaheshwary/reqlog-ui/internal/logger"
	"github.com/sagarmaheshwary/reqlog-ui/internal/reqlog"
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
	params, err := reqlog.ParseParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lines, err := h.reqlogService.Run(c.Request.Context(), params)
	if err != nil {
		h.logger.Error("reqlog run failed", logger.Field{Key: "error", Value: err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"lines": lines})
}

func (h *ReqlogHandler) LogsStream(c *gin.Context) {
	params, err := reqlog.ParseParams(c)
	if err != nil {
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", err.Error())
		c.Writer.Flush()
		return
	}

	// SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // disable nginx buffering

	lineCh := make(chan string, 256)
	errCh, err := h.reqlogService.Stream(c.Request.Context(), params, lineCh)
	if err != nil {
		h.logger.Error("reqlog stream failed", logger.Field{Key: "error", Value: err.Error()})
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", err.Error())
		c.Writer.Flush()
		return
	}

	// Heartbeat ticker keeps the connection alive through proxies.
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	flusher, _ := c.Writer.(http.Flusher)

	if flusher != nil {
		flusher.Flush()
	}

	for {
		select {
		case line, ok := <-lineCh:
			if !ok {
				fmt.Fprintf(c.Writer, "event: done\ndata: stream ended\n\n")
				if flusher != nil {
					flusher.Flush()
				}
				return
			}

			fmt.Fprintf(c.Writer, "data: %s\n\n", sseEscape(line))
			if flusher != nil {
				flusher.Flush()
			}

		case err := <-errCh:
			if err != nil {
				fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", sseEscape(err.Error()))
				if flusher != nil {
					flusher.Flush()
				}
			}
			return

		case <-ticker.C:
			fmt.Fprintf(c.Writer, ": heartbeat\n\n")
			if flusher != nil {
				flusher.Flush()
			}

		case <-c.Request.Context().Done():
			return
		}
	}
}

func sseEscape(s string) string {
	out := ""
	for i, ch := range s {
		if ch == '\n' && i != len(s)-1 {
			out += "\ndata: "
		} else {
			out += string(ch)
		}
	}
	return out
}
