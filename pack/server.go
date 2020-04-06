package pack

import (
	"net/http"

	"github.com/appist/appy/support"
	"go.uber.org/zap"
)

// Server serves the HTTP requests.
type Server struct {
	asset      *support.Asset
	config     *support.Config
	http       *http.Server
	https      *http.Server
	logger     *support.Logger
	middleware []HandlerFunc
	router     *Router
}

// NewServer initializes Server instance.
func NewServer(asset *support.Asset, config *support.Config, logger *support.Logger) *Server {
	router := newRouter()

	httpServer := &http.Server{
		Addr:              config.HTTPHost + ":" + config.HTTPPort,
		Handler:           router,
		MaxHeaderBytes:    config.HTTPMaxHeaderBytes,
		ReadTimeout:       config.HTTPReadTimeout,
		ReadHeaderTimeout: config.HTTPReadHeaderTimeout,
		WriteTimeout:      config.HTTPWriteTimeout,
		IdleTimeout:       config.HTTPIdleTimeout,
	}
	httpServer.ErrorLog = zap.NewStdLog(logger.Desugar())

	httpsServer := &http.Server{
		Addr:              config.HTTPHost + ":" + config.HTTPSSLPort,
		Handler:           router,
		MaxHeaderBytes:    config.HTTPMaxHeaderBytes,
		ReadTimeout:       config.HTTPReadTimeout,
		ReadHeaderTimeout: config.HTTPReadHeaderTimeout,
		WriteTimeout:      config.HTTPWriteTimeout,
		IdleTimeout:       config.HTTPIdleTimeout,
	}
	httpsServer.ErrorLog = zap.NewStdLog(logger.Desugar())

	return &Server{
		asset:      asset,
		config:     config,
		http:       httpServer,
		https:      httpsServer,
		logger:     logger,
		middleware: []HandlerFunc{},
		router:     router,
	}
}
