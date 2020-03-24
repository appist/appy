package handler

import (
	"{{.projectName}}/pkg/app"
	"{{.projectName}}/pkg/handler/middleware"
)

func init() {
	// Setup your HTTP server's global middleware.
	app.Server.Use(middleware.Example())

	// Setup your HTTP server's routes.
	app.Server.GET("/welcome", welcomeIndex)
}
