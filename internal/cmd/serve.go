package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	appyhttp "github.com/appist/appy/internal/http"
	appyorm "github.com/appist/appy/internal/orm"
	appysupport "github.com/appist/appy/internal/support"
)

// NewServeCommand run the HTTP/HTTPS web server without webpack-dev-server.
func NewServeCommand(dbManager *appyorm.DbManager, logger *appysupport.Logger, s *appyhttp.Server) *Command {
	return &Command{
		Use:   "serve",
		Short: "Run the HTTP/HTTPS web server without webpack-dev-server",
		Run: func(cmd *Command, args []string) {
			if appysupport.IsConfigErrored(s.Config(), logger) || appyorm.IsDbManagerErrored(s.Config(), dbManager, logger) {
				os.Exit(-1)
			}

			if s.Config().HTTPSSLEnabled == true && !s.IsSSLCertExisted() {
				logger.Fatal("HTTP_SSL_ENABLED is set to true without SSL certs, please generate using `go run . ssl:setup` first.")
			}

			serve(dbManager, logger, s)
		},
	}
}

func serve(dbManager *appyorm.DbManager, logger *appysupport.Logger, s *appyhttp.Server) {
	httpDone := make(chan bool, 1)
	httpQuit := make(chan os.Signal, 1)
	signal.Notify(httpQuit, os.Interrupt)
	signal.Notify(httpQuit, syscall.SIGTERM)
	go func() {
		<-httpQuit
		logger.Infof("* Gracefully shutting down the server within %s...", s.Config().HTTPGracefulTimeout)
		defer dbManager.CloseAll()

		ctx, cancel := context.WithTimeout(context.Background(), s.Config().HTTPGracefulTimeout)
		defer cancel()

		if err := s.HTTP().Shutdown(ctx); err != nil {
			logger.Fatal(err)
		}

		if s.Config().HTTPSSLEnabled == true {
			if err := s.HTTPS().Shutdown(ctx); err != nil {
				logger.Fatal(err)
			}
		}

		close(httpDone)
	}()

	err := dbManager.ConnectAll(true)
	if err != nil {
		logger.Fatal(err)
	}

	dbManager.PrintInfo()

	for _, info := range s.Info() {
		logger.Info(info)
	}

	go func() {
		if s.Config().HTTPSSLEnabled == true {
			err := s.HTTPS().ListenAndServeTLS(s.Config().HTTPSSLCertPath+"/cert.pem", s.Config().HTTPSSLCertPath+"/key.pem")
			if err != http.ErrServerClosed {
				logger.Fatal(err)
			}
		}
	}()

	err = s.HTTP().ListenAndServe()
	if err != http.ErrServerClosed {
		logger.Fatal(err)
	}

	<-httpDone
}
