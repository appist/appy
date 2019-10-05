package main

import (
	"wewatch/app/handler"

	"github.com/appist/appy"
)

func main() {
	// Setup the application
	appy.Init(assets, nil)

	// Configure routes
	appy.GET("/welcome", handler.WelcomeIndex())

	// Run the application
	appy.Run()
}
