package main

import (
	"{{.projectName}}/pkg/app"

	// Import custom commands.
	_ "{{.projectName}}/cmd"

	// Import database migration/seed.
	_ "{{.projectName}}/db/migrate/primary"
	_ "{{.projectName}}/db/seed/primary"

	// Import HTTP handlers.
	_ "{{.projectName}}/pkg/handler"

	// Import GraphQL handler.
	_ "{{.projectName}}/pkg/graphql"

	// Import background jobs.
	_ "{{.projectName}}/pkg/job"

	// Import mailer with preview.
	_ "{{.projectName}}/pkg/mailer"
)

func main() {
	// Run the application.
	app.Run()
}
