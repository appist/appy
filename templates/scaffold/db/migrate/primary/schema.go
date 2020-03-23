package primary

import (
	"{{.Project.Name}}/pkg/app"
)

func init() {
	db := app.DB("primary")

	if db != nil {
		db.SetSchema(``)
	}
}
