package primary

import (
	"github.com/appist/appy/record"

	"{{.projectName}}/pkg/app"
)

func init() {
	db := app.DB("primary")

	if db != nil {
		db.RegisterSeedTx(
			func(db record.Txer) error {
				return nil
			},
		)
	}
}
