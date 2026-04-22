package server

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sagarmaheshwary/reqlog-ui/internal/config"
	"github.com/sagarmaheshwary/reqlog-ui/internal/logger"
	"github.com/sagarmaheshwary/reqlog-ui/internal/service"
	"github.com/sagarmaheshwary/reqlog-ui/internal/transports/http/server/handler"
	"github.com/sagarmaheshwary/reqlog-ui/internal/transports/http/server/middleware"
)

type Opts struct {
	Config        *config.HTTPServer
	APIKey        string
	Logger        logger.Logger
	ReqlogService service.ReqlogService
}

type HTTPServer struct {
	Config *config.HTTPServer
	Server *http.Server
	Logger logger.Logger
}

func NewServer(opts *Opts) *HTTPServer {
	r := gin.New()

	r.Use(
		gin.Recovery(),
		middleware.ZerologMiddleware(),
	)

	r.StaticFile("/", "./static/index.html")
	r.StaticFile("/login", "./static/login.html")
	r.Static("/static", "./static")

	api := r.Group("/api")

	authHandler := handler.NewAuthHandler(&handler.AuthHandlerOpts{
		APIKey: opts.APIKey,
	})
	api.POST("/auth/token", authHandler.Token)

	protected := api.Group("/")
	protected.Use(middleware.APIKeyAuth(opts.APIKey))
	{
		reqlogHandler := handler.NewReqlogHandler(&handler.ReqlogHandlerOpts{
			ReqlogService: opts.ReqlogService,
			Logger:        opts.Logger,
		})
		protected.GET("/logs", reqlogHandler.Logs)
	}

	return &HTTPServer{
		Config: opts.Config,
		Server: &http.Server{
			Addr:    opts.Config.URL,
			Handler: r,
		},
		Logger: opts.Logger,
	}
}

func (h *HTTPServer) ServeListener(listener net.Listener) error {
	h.Logger.Info("HTTP server started", logger.Field{Key: "address", Value: listener.Addr().String()})
	if err := h.Server.Serve(listener); err != nil && err != http.ErrServerClosed {
		h.Logger.Error("HTTP server failed", logger.Field{Key: "error", Value: err.Error()})
		return err
	}
	return nil
}

func (h *HTTPServer) Serve() error {
	listener, err := net.Listen("tcp", h.Config.URL)
	if err != nil {
		h.Logger.Error("Failed to create HTTP listener",
			logger.Field{Key: "address", Value: h.Config.URL},
			logger.Field{Key: "error", Value: err.Error()},
		)
		return err
	}

	return h.ServeListener(listener)
}
