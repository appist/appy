package record

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/appist/appy/support"
	"github.com/jmoiron/sqlx"
)

// NamedStmt is a prepared statement that executes named queries. Prepare it
// like how you would execute a NamedQuery, but pass in a struct or map when
// executing.
type NamedStmt struct {
	*sqlx.NamedStmt
}

// Stmter implements all Stmt methods and is useful for mocking Stmt in
// unit tests.
type Stmter interface {
	Exec(args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
	Query(args ...interface{}) (*Rows, error)
	QueryContext(ctx context.Context, args ...interface{}) (*Rows, error)
	QueryRow(args ...interface{}) *Row
	QueryRowContext(ctx context.Context, args ...interface{}) *Row
}

// Stmt is a prepared statement.
//
// A Stmt is safe for concurrent use by multiple goroutines.
//
// If a Stmt is prepared on a Tx or Conn, it will be bound to a single underlying
// connection forever. If the Tx or Conn closes, the Stmt will become unusable
// and all operations will return an error. If a Stmt is prepared on a DB, it
// will remain usable for the lifetime of the DB. When the Stmt needs to execute
// on a new underlying connection, it will prepare itself on the new connection
// automatically.
type Stmt struct {
	*sqlx.Stmt
	logger *support.Logger
	query  string
}

// Close closes the statement.
func (s *Stmt) Close() error {
	if s.Stmt != nil {
		return s.Stmt.Close()
	}

	return nil
}

// Exec executes a prepared statement with the given arguments and returns a
// sql.Result summarizing the effect of the statement.
func (s *Stmt) Exec(args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := s.Stmt.Exec(args...)
	s.logger.Info(formatQuery(s.query+formatStmtParams(args...), time.Since(start)))

	return result, err
}

// ExecContext executes a prepared statement with the given arguments and
// returns a sql.Result summarizing the effect of the statement.
func (s *Stmt) ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := s.Stmt.ExecContext(ctx, args...)
	s.logger.Info(formatQuery(s.query+formatStmtParams(args...), time.Since(start)))

	return result, err
}

// Query executes a prepared query statement with the given arguments and
// returns the query results as a *sql.Rows.
func (s *Stmt) Query(args ...interface{}) (*Rows, error) {
	start := time.Now()
	rows, err := s.Stmt.Queryx(args...)
	s.logger.Info(formatQuery(s.query+formatStmtParams(args...), time.Since(start)))

	return &Rows{rows}, err
}

// QueryContext executes a prepared query statement with the given arguments
// and returns the query results as a *sql.Rows.
func (s *Stmt) QueryContext(ctx context.Context, args ...interface{}) (*Rows, error) {
	start := time.Now()
	rows, err := s.Stmt.QueryxContext(ctx, args...)
	s.logger.Info(formatQuery(s.query+formatStmtParams(args...), time.Since(start)))

	return &Rows{rows}, err
}

// QueryRow executes a prepared query statement with the given arguments. If
// an error occurs duringthe execution of the statement, that error will be
// returned by a call to Scan on the returned *sql.Row, which is always non-nil.
// If the query selects no rows, the *sql.Row's Scan will return sql.ErrNoRows.
// Otherwise, the *sql.Row's Scan scans the first selected row and discards the
// rest.
//
// Example usage:
//
//   var name string
//   err := nameByUseridStmt.QueryRow(id).Scan(&name)
func (s *Stmt) QueryRow(args ...interface{}) *Row {
	start := time.Now()
	row := s.Stmt.QueryRowx(args...)
	s.logger.Info(formatQuery(s.query+formatStmtParams(args...), time.Since(start)))

	return &Row{row}
}

// QueryRowContext executes a prepared query statement with the given
// arguments. If an error occurs during the execution of the statement, that
// error will be returned by a call to Scan on the returned *sql.Row, which is
// always non-nil. If the query selects no rows, the *sql.Row's Scan will
// return sql.ErrNoRows. Otherwise, the *sql.Row's Scan scans the first selected
// row and discards the rest.
func (s *Stmt) QueryRowContext(ctx context.Context, args ...interface{}) *Row {
	start := time.Now()
	row := s.Stmt.QueryRowxContext(ctx, args...)
	s.logger.Info(formatQuery(s.query+formatStmtParams(args...), time.Since(start)))

	return &Row{row}
}

func formatStmtParams(args ...interface{}) string {
	params := make([]string, len(args))

	for idx, arg := range args {
		params[idx] = fmt.Sprintf("%+v", arg)
	}

	return " (" + strings.Join(params, ", ") + ")"
}
