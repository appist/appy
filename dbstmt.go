package appy

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// DBStmter implements all DBStmt methods and is useful for mocking DBStmt in unit tests.
type DBStmter interface {
	Exec(args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
	Query(args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error)
	QueryRow(args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row
}

// DBStmt is a prepared statement.
//
// A DBStmt is safe for concurrent use by multiple goroutines.
//
// If a DBStmt is prepared on a DBTx or DBConn, it will be bound to a single underlying connection
// forever. If the DBTx or DBConn closes, the DBStmt will become unusable and all operations will
// return an error. If a DBStmt is prepared on a DB, it will remain usable for the lifetime of the
// DB. When the DBStmt needs to execute on a new underlying connection, it will prepare itself on
// the new connection automatically.
type DBStmt struct {
	*sqlx.Stmt
	logger *Logger
	query  string
}

// Exec executes a prepared statement with the given arguments and returns a Result summarizing
// the effect of the statement.
func (s *DBStmt) Exec(args ...interface{}) (sql.Result, error) {
	s.logger.Info(formatDBQuery(s.query) + formatDBStmtParams(args...))
	return s.Stmt.Exec(args...)
}

// ExecContext executes a prepared statement with the given arguments and returns a Result
// summarizing the effect of the statement.
func (s *DBStmt) ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error) {
	s.logger.Info(formatDBQuery(s.query) + formatDBStmtParams(args...))
	return s.Stmt.ExecContext(ctx, args...)
}

// Query executes a prepared query statement with the given arguments and returns the query
// results as a *Rows.
func (s *DBStmt) Query(args ...interface{}) (*sql.Rows, error) {
	s.logger.Info(formatDBQuery(s.query) + formatDBStmtParams(args...))
	return s.Stmt.Query(args...)
}

// QueryContext executes a prepared query statement with the given arguments and returns the query
// results as a *Rows.
func (s *DBStmt) QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error) {
	s.logger.Info(formatDBQuery(s.query) + formatDBStmtParams(args...))
	return s.Stmt.QueryContext(ctx, args...)
}

// QueryRow executes a prepared query statement with the given arguments. If an error occurs during
// the execution of the statement, that error will be returned by a call to Scan on the returned
// *Row, which is always non-nil. If the query selects no rows, the *Row's Scan will return ErrNoRows.
// Otherwise, the *Row's Scan scans the first selected row and discards the rest.
//
// Example usage:
//
//  var name string
//  err := nameByUseridStmt.QueryRow(id).Scan(&name)
func (s *DBStmt) QueryRow(args ...interface{}) *sql.Row {
	s.logger.Info(formatDBQuery(s.query) + formatDBStmtParams(args...))
	return s.Stmt.QueryRow(args...)
}

// QueryRowContext executes a prepared query statement with the given arguments. If an error occurs
// during the execution of the statement, that error will be returned by a call to Scan on the
// returned *Row, which is always non-nil. If the query selects no rows, the *Row's Scan will
// return ErrNoRows. Otherwise, the *Row's Scan scans the first selected row and discards the rest.
func (s *DBStmt) QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row {
	s.logger.Info(formatDBQuery(s.query) + formatDBStmtParams(args...))
	return s.Stmt.QueryRowContext(ctx, args...)
}

func formatDBStmtParams(args ...interface{}) string {
	params := make([]string, len(args))

	for idx, arg := range args {
		params[idx] = fmt.Sprintf("%+v", arg)
	}

	return " (" + strings.Join(params, ", ") + ")"
}
