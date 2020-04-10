package record

import (
	"fmt"
	"strings"

	"github.com/appist/appy/support"
)

// Engine manages the databases.
type Engine struct {
	databases map[string]DBer
	errors    []error
	logger    *support.Logger
}

// NewEngine initializes the engine instance to manage the databases.
func NewEngine(logger *support.Logger) *Engine {
	engine := &Engine{
		databases: map[string]DBer{},
		logger:    logger,
	}

	dbConfig, errs := parseDBConfig()
	if errs != nil {
		engine.errors = errs
	}

	for name, config := range dbConfig {
		engine.databases[name] = NewDB(config, logger)
	}

	return engine
}

// DB returns the database handle with the specified name.
func (m *Engine) DB(name string) DBer {
	if db, ok := m.databases[name]; ok {
		return db
	}

	return nil
}

// Errors returns the engine errors.
func (m *Engine) Errors() []error {
	return m.errors
}

// Info returns the engine info.
func (m *Engine) Info() string {
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
