package appy

import (
	"fmt"
	"strings"
)

// DBManager manages the application database handlers.
type DBManager struct {
	databases map[string]DBer
	errors    []error
	logger    *Logger
}

// NewDBManager initializes DBManager instance.
func NewDBManager(logger *Logger, support Supporter) *DBManager {
	dbManager := &DBManager{
		databases: map[string]DBer{},
		logger:    logger,
	}

	dbConfig, errs := parseDBConfig(support)
	if errs != nil {
		dbManager.errors = errs
	}

	for name, config := range dbConfig {
		dbManager.databases[name] = NewDB(config, logger, support)
	}

	return dbManager
}

// DB returns the database handle with the specified name.
func (m *DBManager) DB(name string) DBer {
	if db, ok := m.databases[name]; ok {
		return db
	}

	return nil
}

// Errors returns the DBManager errors.
func (m *DBManager) Errors() []error {
	return m.errors
}

// Info returns the DBManager info.
func (m *DBManager) Info() string {
	var dbNames []string
	for name := range m.databases {
		dbNames = append(dbNames, name)
	}

	databases := "none"
	if len(dbNames) > 0 {
		databases = strings.Join(dbNames, ", ")
	}

	return fmt.Sprintf("* DBs: %s", databases)
}
