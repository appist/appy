package appy

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func newServeCommand(dbManager *DbManager, s *Server) *Cmd {
	return &Cmd{
		Use:   "serve",
		Short: "Run the GRPC/HTTP web server",
		Run: func(cmd *Cmd, args []string) {
			CheckConfig(s.config, s.logger)
			CheckDbManager(s.config, dbManager, s.logger)

			if s.config.HTTPSSLEnabled == true && !s.IsSSLCertExisted() {
				s.logger.Fatal("HTTP_SSL_ENABLED is set to true without SSL certs, please generate using `go run . ssl:setup` first.")
			}

			serve(dbManager, s)
		},
	}
}

func serve(dbManager *DbManager, s *Server) {
	httpDone := make(chan bool, 1)
	httpQuit := make(chan os.Signal, 1)
	signal.Notify(httpQuit, os.Interrupt)
	signal.Notify(httpQuit, syscall.SIGTERM)
	go func() {
		<-httpQuit
		s.logger.Infof("* Gracefully shutting down the server within %s...", s.config.HTTPGracefulTimeout)
		defer dbManager.CloseAll()

		ctx, cancel := context.WithTimeout(context.Background(), s.config.HTTPGracefulTimeout)
		defer cancel()

		if err := s.http.Shutdown(ctx); err != nil {
			s.logger.Fatal(err)
		}

		close(httpDone)
	}()

	var err error
	err = dbManager.ConnectAll(true)
	if err != nil {
		s.logger.Fatal(err)
	}

	dbManager.PrintInfo()
	s.PrintInfo()

	if s.config.HTTPSSLEnabled == true {
		err = s.http.ListenAndServeTLS(s.config.HTTPSSLCertPath+"/cert.pem", s.config.HTTPSSLCertPath+"/key.pem")
	} else {
		err = s.http.ListenAndServe()
	}

	if err != http.ErrServerClosed {
		s.logger.Fatal(err)
	}

	<-httpDone
}
