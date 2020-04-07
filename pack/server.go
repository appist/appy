package pack

import (
	"net/http"

	"github.com/appist/appy/support"
	"go.uber.org/zap"
)

// Server processes the HTTP requests.
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

	hs := &http.Server{
		Addr:              config.HTTPHost + ":" + config.HTTPPort,
		Handler:           router,
		MaxHeaderBytes:    config.HTTPMaxHeaderBytes,
		ReadTimeout:       config.HTTPReadTimeout,
		ReadHeaderTimeout: config.HTTPReadHeaderTimeout,
		WriteTimeout:      config.HTTPWriteTimeout,
		IdleTimeout:       config.HTTPIdleTimeout,
	}
	hs.ErrorLog = zap.NewStdLog(logger.Desugar())

	hss := &http.Server{
		Addr:              config.HTTPHost + ":" + config.HTTPSSLPort,
		Handler:           router,
		MaxHeaderBytes:    config.HTTPMaxHeaderBytes,
		ReadTimeout:       config.HTTPReadTimeout,
		ReadHeaderTimeout: config.HTTPReadHeaderTimeout,
		WriteTimeout:      config.HTTPWriteTimeout,
		IdleTimeout:       config.HTTPIdleTimeout,
	}
	hss.ErrorLog = zap.NewStdLog(logger.Desugar())

	return &Server{
		asset:      asset,
		config:     config,
		http:       hs,
		https:      hss,
		logger:     logger,
		middleware: []HandlerFunc{},
		router:     router,
	}
}
