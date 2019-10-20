package main

import (
	"github.com/appist/appy"

	"wewatch/app"
	"wewatch/app/handler"
)

func main() {
	// Setup the app instance.
	appy.Init(assets, app.Config, nil)

	// Configure the application routes.
	appy.GET("/welcome", handler.WelcomeIndex())

	// Run the application.
	appy.Run()
}
