package primary

import (
	"github.com/appist/appy"

	"{{.Project.Name}}/pkg/app"
)

func init() {
	db := app.DB("primary")

	if db != nil {
		db.RegisterSeedTx(
			func(db *appy.DBTx) error {
				return nil
			},
		)
	}
}
