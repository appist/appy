package main

import (
	"github.com/appist/appy"

	"wewatch/app/handler"

	_ "wewatch/app"
	_ "wewatch/db/migrations/primary"
)

func main() {
	// Configure the application routes.
	appy.GET("/welcome", handler.WelcomeIndex())

	// Run the application.
	appy.Run()
}
