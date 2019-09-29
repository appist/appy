package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	ahttp "github.com/appist/appy/http"
	"github.com/spf13/cobra"
)

// NewHTTPServeCommand runs the HTTP/HTTPS web server.
func NewHTTPServeCommand(s *ahttp.ServerT) *cobra.Command {
	return &cobra.Command{
		Use:   "http:serve",
		Short: "Run the HTTP/HTTPS web server.",
		Run: func(cmd *cobra.Command, args []string) {
			checkSSLCerts()
			s.SetupCSR()
			serveHTTPServer(s)
		},
	}
}

func serveHTTPServer(s *ahttp.ServerT) {
	logServerInfo(config.HTTPSSLEnabled, config.HTTPPort, config.HTTPSSLPort)

	httpDone := make(chan bool, 1)
	httpQuit := make(chan os.Signal, 1)
	signal.Notify(httpQuit, os.Interrupt)
	signal.Notify(httpQuit, syscall.SIGTERM)
	go func() {
		<-httpQuit
		logger.Infof("Gracefully shutting down the server within %s...", config.HTTPGracefulTimeout)
		ctx, cancel := context.WithTimeout(context.Background(), config.HTTPGracefulTimeout)
		defer cancel()

		if config.HTTPSSLEnabled == true {
			if err := s.HTTPS.Shutdown(ctx); err != nil {
				logger.Fatal(err)
			}
		} else {
			if err := s.HTTP.Shutdown(ctx); err != nil {
				logger.Fatal(err)
			}
		}

		close(httpDone)
	}()

	if config.HTTPSSLEnabled == true {
		if err := s.HTTPS.ListenAndServeTLS(config.HTTPSSLCertPath+"/cert.pem", config.HTTPSSLCertPath+"/key.pem"); err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	} else {
		if err := s.HTTP.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}

	<-httpDone
}
