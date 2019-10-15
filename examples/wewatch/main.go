package main

import (
	"github.com/appist/appy"

	"wewatch/app/handler"
)

func main() {
	// Setup the app instance.
	appy.Init(assets, nil)

	// // Configure routes
	appy.GET("/welcome", handler.WelcomeIndex())

	// Run the application
	appy.Run()
}
