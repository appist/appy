package job

import (
	"{{.Project.Name}}/pkg/app"
	"{{.Project.Name}}/pkg/job/middleware"
)

func init() {
	// Setup your worker's global middleware.
	app.Worker.Use(middleware.Example)
}
