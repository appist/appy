package record

import (
	"strings"
	"sync"

	"github.com/appist/appy/support"
	"github.com/jmoiron/sqlx"
)

const (
	loggerDBPrefix = "[DB] "
)

type (
	// DBer implements all DB methods.
	DBer interface {
	}

	// DB manages the database config/connection/migrations.
	DB struct {
		*sqlx.DB
		config     *Config
		logger     *support.Logger
		migrations []*Migration
		mu         *sync.Mutex
		schema     string
		seed       func(*Tx) error
	}

	// Row is a wrapper around sqlx.Row.
	Row struct {
		*sqlx.Row
	}

	// Rows is a wrapper around sqlx.Rows.
	Rows struct {
		*sqlx.Rows
	}
)

// NewDB initializes the database handler that is used to connect to the database.
func NewDB(config *Config, logger *support.Logger) *DB {
	return &DB{
		nil,
		config,
		logger,
		nil,
		&sync.Mutex{},
		"",
		nil,
	}
}

func formatQuery(query string) string {
	formattedQuery := strings.Trim(query, "\n")
	formattedQuery = strings.TrimSpace(formattedQuery)

	if strings.Contains(formattedQuery, "\n") {
		formattedQuery = strings.ReplaceAll(formattedQuery, "\n", "\n\t\t\t\t\t     ")
	}

	return loggerDBPrefix + formattedQuery
}
