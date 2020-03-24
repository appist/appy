package primary

import (
	"github.com/appist/appy"

	"{{.projectName}}/pkg/app"
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
