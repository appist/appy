package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/appist/appy/pack"
	"github.com/appist/appy/record"
	"github.com/appist/appy/support"
)

func newServeCommand(dbManager *record.Engine, logger *support.Logger, server *pack.Server) *Command {
	return &Command{
		Use:   "serve",
		Short: "Run the HTTP/HTTPS web server without `webpack-dev-server`",
		Run: func(cmd *Command, args []string) {
			if len(server.Config().Errors()) > 0 {
				logger.Fatal(server.Config().Errors()[0])
			}

			if len(dbManager.Errors()) > 0 {
				logger.Fatal(dbManager.Errors()[0])
			}

			if server.Config().HTTPSSLEnabled && !server.IsSSLCertExisted() {
				logger.Fatal("HTTP_SSL_ENABLED is set to true without SSL certs, please generate using `go run . ssl:setup` first.")
			}

			serve(dbManager, logger, server)
		},
	}
}

func serve(dbManager *record.Engine, logger *support.Logger, server *pack.Server) {
	httpDone := make(chan bool, 1)
	httpQuit := make(chan os.Signal, 1)
	signal.Notify(httpQuit, os.Interrupt)
	signal.Notify(httpQuit, syscall.SIGTERM)

	go func() {
		<-httpQuit
		logger.Infof("* Gracefully shutting down the server within %s...", server.Config().HTTPGracefulShutdownTimeout)

		for _, db := range dbManager.Databases() {
			err := db.Close()
			if err != nil {
				logger.Fatal(err)
			}
		}

		// TODO: Allow graceful handling from the app.

		ctx, cancel := context.WithTimeout(context.Background(), server.Config().HTTPGracefulShutdownTimeout)
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

	for _, db := range dbManager.Databases() {
		err := db.Connect()
		if err != nil {
			logger.Fatal(err)
		}
	}

	for _, info := range server.Info() {
		if strings.Contains(info, "* Listening on") {
			logger.Info(dbManager.Info())
		}

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
