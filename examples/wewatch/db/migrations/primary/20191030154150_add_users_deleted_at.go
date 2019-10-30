package primary

import (
	"github.com/appist/appy"
)

func init() {
	appy.Db["primary"].RegisterMigrationTx(
		// Up migration
		func(db *appy.AppDbHandlerTx) error {
			_, err := db.Exec(`
				ALTER TABLE users
					ADD COLUMN deleted_at TIMESTAMPTZ
			`)
			return err
		},

		// Down migration
		func(db *appy.AppDbHandlerTx) error {
			_, err := db.Exec(`
				ALTER TABLE users
					DROP COLUMN IF EXISTS deleted_at
			`)
			return err
		},
	)
}
