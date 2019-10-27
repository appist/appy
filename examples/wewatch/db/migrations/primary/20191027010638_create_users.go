package primary

import (
	"github.com/appist/appy"
)

func init() {
	appy.Db["primary"].RegisterMigrationTx(
		// Up migration
		func(db *appy.AppDbHandlerTx) error {
			return nil
		},

		// Down migration
		func(db *appy.AppDbHandlerTx) error {
			return nil
		},
	)
}
