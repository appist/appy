package http

import (
	"context"
	"net/http"
	"time"

	"github.com/appist/appy/o11y"
	"github.com/appist/appy/support"
	"go.uber.org/zap"
)

// ServerConfig indicates how a HTTP server should be initialised.
type ServerConfig struct {
	// Address is the HTTP address to listen on.
	Address string

	// PreShutdownHandler is a function that runs before the HTTP server is gracefully shut down.
	PreShutdownHandler func() error

	// ShutdownTimeout indicates the duration to wait before completely shutting down the server.
	// By default, it is 10 * time.Second.
	ShutdownTimeout time.Duration

	// Handler is the router.
	Handler http.Handler

	// IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are
	// enabled. If IdleTimeout is zero, the value of ReadTimeout is used. If both are zero, there is
	// no timeout. By default, it is 75 * time.Second.
	IdleTimeout time.Duration

	// MaxHeaderBytes controls the maximum number of bytes the server will read parsing the request
	// header's keys and values, including the request line. It does not limit the size of the
	// request body. If zero, DefaultMaxHeaderBytes (1 << 20 which is 1 MB) is used. By default, it
	// is 0.
	MaxHeaderBytes int

	// ReadTimeout is the maximum duration for reading the entire request, including the body. A zero
	// or negative value means there will be no timeout. Because ReadTimeout does not let Handlers
	// make per-request decisions on each request body's acceptable deadline or upload rate, most
	// users will prefer to use ReadHeaderTimeout. It is valid to use them both. By default, it is
	// 60 * time.Second.
	ReadTimeout time.Duration

	// ReadHeaderTimeout is the amount of time allowed to read request headers. The connection's read
	// deadline is reset after reading the headers and the Handler can decide what is considered too
	// slow for the body. If ReadHeaderTimeout is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout. By default, it is 60 * time.Second.
	ReadHeaderTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response. It is reset
	// whenever a new request's header is read. Like ReadTimeout, it does not let Handlers make
	// decisions on a per-request basis. A zero or negative value means there will be no timeout.
	// By default, it is 60 * time.Second.
	WriteTimeout time.Duration

	// PostServeHandler is the handler function to trigger after the server starts serving.
	PostServeHandler func() error

	// TracerProvider is the provider that uses the exporter to push traces to the collector.
	TracerProvider *o11y.TracerProvider
}

// Server wraps the *http.Server.
type Server struct {
	config *ServerConfig
	logger *support.Logger
	server *http.Server
}

// NewServer initialises a HTTP server instance.
func NewServer(cfg *ServerConfig, logger *support.Logger) *Server {
	defaultServerConfig(cfg)

	return &Server{
		cfg,
		logger,
		&http.Server{
			Addr:              cfg.Address,
			ErrorLog:          zap.NewStdLog(logger.Desugar()),
			Handler:           cfg.Handler,
			IdleTimeout:       cfg.IdleTimeout,
			MaxHeaderBytes:    cfg.MaxHeaderBytes,
			ReadTimeout:       cfg.ReadTimeout,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			WriteTimeout:      cfg.WriteTimeout,
		},
	}
}

// Addr returns the server's network address.
func (s *Server) Addr() string {
	return s.config.Address
}

// PreShutdownHandler is a function that runs before the HTTP server is gracefully shut down.
func (s *Server) PreShutdownHandler() error {
	return s.config.PreShutdownHandler()
}

// Serve listens on the TCP network address srv.Addr and then calls Serve to handle
// requests on incoming connections. Accepted connections are configured to enable TCP
// keep-alives.
func (s *Server) Serve() error {
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Shutdown stops the HTTP server gracefully. It stops the server from accepting new connections
// and HTTP requests and blocks until all the pending HTTP requests are finished.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// PostServeHandler returns the handler function to trigger after the server starts serving.
func (s *Server) PostServeHandler() error {
	return s.config.PostServeHandler()
}

func defaultServerConfig(c *ServerConfig) {
	if c.PreShutdownHandler == nil {
		c.PreShutdownHandler = func() error {
			return nil
		}
	}

	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = 10 * time.Second
	}

	if c.IdleTimeout == 0 {
		c.IdleTimeout = 75 * time.Second
	}

	if c.PostServeHandler == nil {
		c.PostServeHandler = func() error {
			return nil
		}
	}

	if c.ReadHeaderTimeout == 0 {
		c.ReadHeaderTimeout = 60 * time.Second
	}

	if c.ReadTimeout == 0 {
		c.ReadTimeout = 60 * time.Second
	}

	if c.WriteTimeout == 0 {
		c.WriteTimeout = 60 * time.Second
	}
}
