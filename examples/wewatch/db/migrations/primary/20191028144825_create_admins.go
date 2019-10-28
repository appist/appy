package primary

import (
	"github.com/appist/appy"
)

func init() {
	appy.Db["primary"].RegisterMigrationTx(
		// Up migration
		func(db *appy.AppDbHandlerTx) error {
			_, err := db.Exec(`
				CREATE TABLE IF NOT EXISTS admins (
					id serial PRIMARY KEY,
					email VARCHAR(355) UNIQUE NOT NULL,
					created_at TIMESTAMPTZ NOT NULL,
					updated_at TIMESTAMPTZ
				)`)
			return err
		},

		// Down migration
		func(db *appy.AppDbHandlerTx) error {
			_, err := db.Exec(`DROP TABLE IF EXISTS admins`)
			return err
		},
	)
}
