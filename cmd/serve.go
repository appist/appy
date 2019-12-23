package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	ah "github.com/appist/appy/http"
	"github.com/appist/appy/support"
)

// NewServeCommand run the HTTP/HTTPS web server without webpack-dev-server.
func NewServeCommand(logger *support.Logger, server *ah.Server) *Command {
	return &Command{
		Use:   "serve",
		Short: "Run the HTTP/HTTPS web server without webpack-dev-server",
		Run: func(cmd *Command, args []string) {
			if support.IsConfigErrored(server.Config(), logger) {
				os.Exit(-1)
			}

			if server.Config().HTTPSSLEnabled == true && !server.IsSSLCertExisted() {
				logger.Fatal("HTTP_SSL_ENABLED is set to true without SSL certs, please generate using `go run . ssl:setup` first.")
			}

			serve(logger, server)
		},
	}
}

func serve(logger *support.Logger, server *ah.Server) {
	httpDone := make(chan bool, 1)
	httpQuit := make(chan os.Signal, 1)
	signal.Notify(httpQuit, os.Interrupt)
	signal.Notify(httpQuit, syscall.SIGTERM)
	go func() {
		<-httpQuit
		logger.Infof("* Gracefully shutting down the server within %s...", server.Config().HTTPGracefulTimeout)

		ctx, cancel := context.WithTimeout(context.Background(), server.Config().HTTPGracefulTimeout)
		defer cancel()

		if err := server.HTTP().Shutdown(ctx); err != nil {
			logger.Fatal(err)
		}

		if server.Config().HTTPSSLEnabled == true {
			if err := server.HTTPS().Shutdown(ctx); err != nil {
				logger.Fatal(err)
			}
		}

		close(httpDone)
	}()

	for _, info := range server.Info() {
		logger.Info(info)
	}

	go func() {
		if server.Config().HTTPSSLEnabled == true {
			err := server.HTTPS().ListenAndServeTLS(server.Config().HTTPSSLCertPath+"/cert.pem", server.Config().HTTPSSLCertPath+"/key.pem")
			if err != http.ErrServerClosed {
				logger.Fatal(err)
			}
		}
	}()

	err := server.HTTP().ListenAndServe()
	if err != http.ErrServerClosed {
		logger.Fatal(err)
	}

	<-httpDone
}
