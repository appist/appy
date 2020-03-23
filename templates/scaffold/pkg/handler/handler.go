package handler

import (
	"{{.Project.Name}}/pkg/app"
	"{{.Project.Name}}/pkg/handler/middleware"
)

func init() {
	// Setup your HTTP server's global middleware.
	app.Server.Use(middleware.Example())

	// Setup your HTTP server's routes.
	app.Server.GET("/welcome", welcomeIndex)
}
