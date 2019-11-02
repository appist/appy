package appy

import (
	"os"
	"strings"
	"sync"
)

type (
	DbManagerer interface {
	}

	// DbManager manages multiple databases.
	DbManager struct {
		dbs    map[string]*Db
		errors []error
		logger *Logger
		mu     sync.Mutex
	}
)

// NewDbManager initializes DbManager instance.
func NewDbManager(logger *Logger) *DbManager {
	dbManager := &DbManager{
		dbs:    map[string]*Db{},
		logger: logger,
	}
	dbConfig, errs := parseDbConfig()
	if errs != nil {
		dbManager.errors = errs
	}

	for name, val := range dbConfig {
		dbManager.dbs[name] = &Db{
			config: val,
			logger: logger,
		}
	}

	return dbManager
}

// ConnectAll establishes connections to all the databases.
func (m *DbManager) ConnectAll(sameDb bool) error {
	for _, db := range m.dbs {
		err := db.Connect(sameDb)
		if err != nil {
			return err
		}
	}

	return nil
}

// CloseAll closes connections to all the databases.
func (m *DbManager) CloseAll() error {
	for _, db := range m.dbs {
		err := db.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// Db returns the database instance with the specified name.
func (m *DbManager) Db(name string) *Db {
	if db, ok := m.dbs[name]; ok {
		return db
	}

	return nil
}

// PrintInfo prints the database manager info.
func (m *DbManager) PrintInfo() {
	var dbNames []string
	for name := range m.dbs {
		dbNames = append(dbNames, name)
	}

	dbs := "none"
	if len(dbNames) > 0 {
		dbs = strings.Join(dbNames, ", ")
	}

	m.logger.Infof("* Available DBs: %s", dbs)
}

// CheckDbManager is used to check if DB manager contains any error during initialization.
func CheckDbManager(config *Config, dbManager *DbManager, logger *Logger) {
	if dbManager != nil && dbManager.errors != nil {
		for _, err := range dbManager.errors {
			logger.Info(err.Error())
		}

		os.Exit(-1)
	}
}
