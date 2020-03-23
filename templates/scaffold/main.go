package main

import (
	"{{.Project.Name}}/pkg/app"

	// Import custom commands.
	_ "{{.Project.Name}}/cmd"

	// Import database migration/seed.
	_ "{{.Project.Name}}/db/migrate/primary"

	// Import GraphQL handler.
	_ "{{.Project.Name}}/pkg/graphql"

	// Import HTTP handlers.
	_ "{{.Project.Name}}/pkg/handler"

	// Import background jobs.
	_ "{{.Project.Name}}/pkg/job"

	// Import mailer with preview.
	_ "{{.Project.Name}}/pkg/mailer"
)

func main() {
	// Run the application.
	app.Run()
}
