//+build !test

package appy

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func newServeCommand(logger *Logger, server *Server) *Command {
	return &Command{
		Use:   "serve",
		Short: "Run the HTTP/HTTPS web server without `webpack-dev-server`",
		Run: func(cmd *Command, args []string) {
			if len(server.Config().Errors()) > 0 {
				logger.Fatal(server.Config().Errors()[0])
			}

			if server.Config().HTTPSSLEnabled && !server.IsSSLCertExisted() {
				logger.Fatal("HTTP_SSL_ENABLED is set to true without SSL certs, please generate using `go run . ssl:setup` first.")
			}

			serve(logger, server)
		},
	}
}

func serve(logger *Logger, server *Server) {
	httpDone := make(chan bool, 1)
	httpQuit := make(chan os.Signal, 1)
	signal.Notify(httpQuit, os.Interrupt)
	signal.Notify(httpQuit, syscall.SIGTERM)
	go func() {
		<-httpQuit
		logger.Infof("* Gracefully shutting down the server within %s...", server.Config().HTTPGracefulTimeout)

		// TODO: Allow graceful handling from the app.

		ctx, cancel := context.WithTimeout(context.Background(), server.Config().HTTPGracefulTimeout)
		defer cancel()

		if err := server.HTTP().Shutdown(ctx); err != nil {
			logger.Fatal(err)
		}

		if server.Config().HTTPSSLEnabled {
			if err := server.HTTPS().Shutdown(ctx); err != nil {
				logger.Fatal(err)
			}
		}

		close(httpDone)
	}()

	// TODO: Add database connection establishing

	for _, info := range server.Info() {
		logger.Info(info)
	}

	go func() {
		if server.Config().HTTPSSLEnabled {
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
