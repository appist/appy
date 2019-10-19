package main

import (
	"github.com/appist/appy"

	"wewatch/app/handler"
)

type config struct {
	AppName string `env:"APP_NAME"`
}

func main() {
	cfg := &config{}

	// Setup the app instance.
	appy.Init(assets, cfg, nil)

	// // Configure routes
	appy.GET("/welcome", handler.WelcomeIndex())

	// Run the application
	appy.Run()
}
