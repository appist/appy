package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	ah "appist/appy/http"

	"github.com/spf13/cobra"
)

// NewHTTPServeCommand runs the HTTP/HTTPS web server.
func NewHTTPServeCommand(s *ah.ServerT) *cobra.Command {
	return &cobra.Command{
		Use:   "http:serve",
		Short: "Run the HTTP/HTTPS web server.",
		Run: func(cmd *cobra.Command, args []string) {
			checkSSLCerts(s)
			s.SetupCSR()
			serveHTTPServer(s)
		},
	}
}

func serveHTTPServer(s *ah.ServerT) {
	logServerInfo(s)

	httpDone := make(chan bool, 1)
	httpQuit := make(chan os.Signal, 1)
	signal.Notify(httpQuit, os.Interrupt)
	signal.Notify(httpQuit, syscall.SIGTERM)
	go func() {
		<-httpQuit
		logger.Infof("Gracefully shutting down the server within %s...", s.Config.HTTPGracefulTimeout)
		ctx, cancel := context.WithTimeout(context.Background(), s.Config.HTTPGracefulTimeout)
		defer cancel()

		if err := s.HTTP.Shutdown(ctx); err != nil {
			logger.Fatal(err)
		}

		close(httpDone)
	}()

	var err error
	if s.Config.HTTPSSLEnabled == true {
		err = s.HTTP.ListenAndServeTLS(s.Config.HTTPSSLCertPath+"/cert.pem", s.Config.HTTPSSLCertPath+"/key.pem")
	} else {
		err = s.HTTP.ListenAndServe()
	}

	if err != http.ErrServerClosed {
		logger.Fatal(err)
	}

	<-httpDone
}
