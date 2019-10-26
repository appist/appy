package app

import (
	"github.com/appist/appy"

	"wewatch/app/handler"
)

type config struct {
	AppName string `env:"APP_NAME"`
}

// Config is the application config.
var Config *config

func init() {
	Config = &config{}

	// Setup the app instance.
	appy.Init(assets, Config, nil)

	// Configure the application routes.
	appy.GET("/welcome", handler.WelcomeIndex())
}
