package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	ah "github.com/appist/appy/http"

	"github.com/spf13/cobra"
)

// NewServeCommand runs the HTTP/HTTPS web server.
func NewServeCommand(s *ah.ServerT) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run the GRPC/HTTP web server.",
		Run: func(cmd *cobra.Command, args []string) {
			s.CheckSSLCerts()
			// s.SetupCSR()
			serve(s)
		},
	}
}

func serve(s *ah.ServerT) {
	s.PrintInfo()

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
