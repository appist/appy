package main

import (
	"github.com/appist/appy"

	_ "wewatch/app"
	_ "wewatch/db/migrations/primary"
)

func main() {
	// Run the application.
	appy.Run()
}
