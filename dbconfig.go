package appy

import "github.com/go-pg/pg/v9"

// DBConfig contains database connection options.
type DBConfig struct {
	pg.Options
	Replica               bool
	SchemaSearchPath      string
	SchemaMigrationsTable string
}
