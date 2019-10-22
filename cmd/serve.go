package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/appist/appy/core"
)

// NewServeCommand runs the GRPC/HTTP web server.
func NewServeCommand(s core.AppServer) *AppCmd {
	return &AppCmd{
		Use:   "serve",
		Short: "Runs the GRPC/HTTP web server.",
		Run: func(cmd *AppCmd, args []string) {
			if s.Config.HTTPSSLEnabled == true && !s.IsSSLCertsExist() {
				s.Logger.Fatal("HTTP_SSL_ENABLED is set to true without SSL certs, please generate using `go run . ssl:setup` first.")
			}

			serve(s)
		},
	}
}

func serve(s core.AppServer) {
	s.PrintInfo()

	httpDone := make(chan bool, 1)
	httpQuit := make(chan os.Signal, 1)
	signal.Notify(httpQuit, os.Interrupt)
	signal.Notify(httpQuit, syscall.SIGTERM)
	go func() {
		<-httpQuit
		s.Logger.Infof("* Gracefully shutting down the server within %s...\n", s.Config.HTTPGracefulTimeout)
		ctx, cancel := context.WithTimeout(context.Background(), s.Config.HTTPGracefulTimeout)
		defer cancel()

		if err := s.HTTP.Shutdown(ctx); err != nil {
			s.Logger.Fatal(err)
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
		s.Logger.Fatal(err)
	}

	<-httpDone
}
