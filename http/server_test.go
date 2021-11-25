package http

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/appist/appy/o11y"
	"github.com/appist/appy/support"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	t.Run("should return a server instance with default configuration when the configuration is empty", func(t *testing.T) {
		logger, _, _ := support.NewTestLogger()
		server := NewServer(&ServerConfig{}, logger)

		assert.NotNil(t, server.logger)
		assert.Empty(t, server.config.Address)
		assert.Empty(t, server.Addr())
		assert.Nil(t, server.config.PreShutdownHandler())
		assert.Nil(t, server.PreShutdownHandler())
		assert.Nil(t, server.Shutdown())
		assert.Equal(t, 10*time.Second, server.config.ShutdownTimeout)
		assert.Nil(t, server.config.Handler)
		assert.Equal(t, 75*time.Second, server.config.IdleTimeout)
		assert.Equal(t, 0, server.config.MaxHeaderBytes)
		assert.Equal(t, 60*time.Second, server.config.ReadHeaderTimeout)
		assert.Equal(t, 60*time.Second, server.config.ReadTimeout)
		assert.Nil(t, server.config.TracerProvider)
		assert.Equal(t, 60*time.Second, server.config.WriteTimeout)
	})

	t.Run("should return a server instance with overriden configuration when the configuration is not empty", func(t *testing.T) {
		logger, _, _ := support.NewTestLogger()
		server := NewServer(&ServerConfig{
			Address:         "0.0.0.0:3000",
			ShutdownTimeout: 30 * time.Second,
			PreShutdownHandler: func() error {
				return errors.New("PreShutdownHandler error")
			},
			Handler:           http.DefaultServeMux,
			IdleTimeout:       120 * time.Second,
			MaxHeaderBytes:    120,
			ReadHeaderTimeout: 120 * time.Second,
			ReadTimeout:       120 * time.Second,
			PostServeHandler: func() error {
				return errors.New("PostServeHandler error")
			},
			TracerProvider: &o11y.TracerProvider{},
			WriteTimeout:   120 * time.Second,
		}, logger)

		assert.NotNil(t, server.logger)
		assert.Equal(t, "0.0.0.0:3000", server.config.Address)
		assert.Equal(t, "0.0.0.0:3000", server.Addr())
		assert.EqualError(t, server.config.PreShutdownHandler(), "PreShutdownHandler error")
		assert.EqualError(t, server.PreShutdownHandler(), "PreShutdownHandler error")
		assert.Nil(t, server.Shutdown())
		assert.Equal(t, 30*time.Second, server.config.ShutdownTimeout)
		assert.NotNil(t, server.config.Handler)
		assert.Equal(t, 120*time.Second, server.config.IdleTimeout)
		assert.Equal(t, 120, server.config.MaxHeaderBytes)
		assert.EqualError(t, server.PostServeHandler(), "PostServeHandler error")
		assert.Equal(t, 120*time.Second, server.config.ReadHeaderTimeout)
		assert.Equal(t, 120*time.Second, server.config.ReadTimeout)
		assert.NotNil(t, server.config.TracerProvider)
		assert.Equal(t, 120*time.Second, server.config.WriteTimeout)
	})
}
