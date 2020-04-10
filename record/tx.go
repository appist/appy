package record

import (
	"context"
	"database/sql"

	"github.com/appist/appy/support"
	"github.com/jmoiron/sqlx"
)

// Txer implements all Tx methods and is useful for mocking Tx in unit
// tests.
type Txer interface {
	Commit() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*Stmt, error)
	PrepareContext(ctx context.Context, query string) (*Stmt, error)
	Query(query string, args ...interface{}) (*Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error)
	QueryRow(query string, args ...interface{}) *Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *Row
	Rollback() error
	Stmt(stmt *Stmt) *Stmt
	StmtContext(ctx context.Context, stmt *Stmt) *Stmt
}

// Tx is an in-progress database transaction.
//
// A transaction must end with a call to Commit or Rollback.
//
// After a call to Commit or Rollback, all operations on the transaction fail
// with ErrTxDone.
//
// The statements prepared for a transaction by calling the transaction's
// Prepare or Stmt methods are closed by the call to Commit or Rollback.
type Tx struct {
	*sqlx.Tx
	logger *support.Logger
}

// Commit commits the transaction.
func (tx *Tx) Commit() error {
	tx.logger.Info(formatQuery("COMMIT;"))
	return tx.Tx.Commit()
}

// Exec executes a query that doesn't return rows. For example: an INSERT and
// UPDATE.
func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	tx.logger.Infof(formatQuery(query), args...)
	return tx.Tx.Exec(query, args...)
}

// ExecContext executes a query that doesn't return rows. For example: an INSERT
// and UPDATE.
func (tx *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	tx.logger.Infof(formatQuery(query), args...)
	return tx.Tx.ExecContext(ctx, query, args...)
}

// Prepare creates a prepared statement for use within a transaction.
//
// The returned statement operates within the transaction and can no longer be
// used once the transaction has been committed or rolled back.
//
// To use an existing prepared statement on this transaction, see Tx.Stmt.
func (tx *Tx) Prepare(query string) (*Stmt, error) {
	stmt, err := tx.Tx.Preparex(query)
	return &Stmt{stmt, tx.logger, query}, err
}

// PrepareContext creates a prepared statement for use within a transaction.
//
// The returned statement operates within the transaction and will be closed
// when the transaction has been committed or rolled back.
//
// To use an existing prepared statement on this transaction, see Tx.Stmt.
//
// The provided context will be used for the preparation of the context, not
// for the execution of the returned statement. The returned statement will
// run in the transaction context.
func (tx *Tx) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	stmt, err := tx.Tx.PreparexContext(ctx, query)
	return &Stmt{stmt, tx.logger, query}, err
}

// Query executes a query that returns rows, typically a SELECT.
func (tx *Tx) Query(query string, args ...interface{}) (*Rows, error) {
	tx.logger.Infof(formatQuery(query), args...)

	rows, err := tx.Tx.Queryx(query, args...)
	return &Rows{rows}, err
}

// QueryContext executes a query that returns rows, typically a SELECT.
func (tx *Tx) QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error) {
	tx.logger.Infof(formatQuery(query), args...)

	rows, err := tx.Tx.QueryxContext(ctx, query, args...)
	return &Rows{rows}, err
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until
// sql.Row's Scan method is called. If the query selects no rows, the
// *sql.Row's Scan will return sql.ErrNoRows. Otherwise, the *sql.Row's
// Scan scans the first selected row and discards the rest.
func (tx *Tx) QueryRow(query string, args ...interface{}) *Row {
	tx.logger.Infof(formatQuery(query), args...)
	return &Row{tx.Tx.QueryRowx(query, args...)}
}

// QueryRowContext executes a query that is expected to return at most
// one row. QueryRowContext always returns a non-nil value. Errors are deferred
// until sql.Row's Scan method is called. If the query selects no rows, the
// *sql.Row's Scan will return sql.ErrNoRows. Otherwise, the *sql.Row's Scan
// scans the first selected row and discards the rest.
func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *Row {
	tx.logger.Infof(formatQuery(query), args...)
	return &Row{tx.Tx.QueryRowxContext(ctx, query, args...)}
}

// Rollback aborts the transaction.
func (tx *Tx) Rollback() error {
	tx.logger.Info(formatQuery("ROLLBACK;"))
	return tx.Tx.Rollback()
}

// Stmt returns a transaction-specific prepared statement from an existing
// statement.
//
// Example:
//  updateMoney, err := db.Prepare("UPDATE balance SET money=money+? WHERE id=?")
//  ...
//  tx, err := db.Begin()
//  ...
//  res, err := tx.Stmt(updateMoney).Exec(123.45, 98293203)
//
// The returned statement operates within the transaction and will be closed
// when the transaction has been committed or rolled back.
func (tx *Tx) Stmt(stmt *Stmt) *Stmt {
	return &Stmt{tx.Tx.Stmtx(stmt.Stmt), tx.logger, stmt.query}
}

// StmtContext returns a transaction-specific prepared statement from an
// existing statement.
//
// Example:
//  updateMoney, err := db.Prepare("UPDATE balance SET money=money+? WHERE id=?")
//  ...
//  tx, err := db.Begin()
//  ...
//  res, err := tx.StmtContext(ctx, updateMoney).Exec(123.45, 98293203)
//
// The provided context is used for the preparation of the statement, not for
// the execution of the statement.
//
// The returned statement operates within the transaction and will be closed
// when the transaction has been committed or rolled back.
func (tx *Tx) StmtContext(ctx context.Context, stmt *Stmt) *Stmt {
	return &Stmt{tx.Tx.StmtxContext(ctx, stmt.Stmt), tx.logger, stmt.query}
}
