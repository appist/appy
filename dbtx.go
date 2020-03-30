package appy

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// DBTxer implements all DBTx methods and is useful for mocking DBTx in unit tests.
type DBTxer interface {
	Commit() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*DBStmt, error)
	PrepareContext(ctx context.Context, query string) (*DBStmt, error)
	Query(query string, args ...interface{}) (*DBRows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*DBRows, error)
	QueryRow(query string, args ...interface{}) *DBRow
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *DBRow
	Rollback() error
	Stmt(stmt *DBStmt) *DBStmt
	StmtContext(ctx context.Context, stmt *DBStmt) *DBStmt
}

// DBTx is an in-progress database transaction.
//
// A transaction must end with a call to Commit or Rollback.
//
// After a call to Commit or Rollback, all operations on the transaction fail with ErrTxDone.
//
// The statements prepared for a transaction by calling the transaction's Prepare or Stmt methods
// are closed by the call to Commit or Rollback.
type DBTx struct {
	*sqlx.Tx
	logger *Logger
}

// Commit commits the transaction.
func (tx *DBTx) Commit() error {
	tx.logger.Info(formatDBQuery("COMMIT;"))
	return tx.Tx.Commit()
}

// Exec executes a query that doesn't return rows. For example: an INSERT and UPDATE.
func (tx *DBTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	tx.logger.Infof(formatDBQuery(query), args...)
	return tx.Tx.Exec(query, args...)
}

// ExecContext executes a query that doesn't return rows. For example: an INSERT and UPDATE.
func (tx *DBTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	tx.logger.Infof(formatDBQuery(query), args...)
	return tx.Tx.ExecContext(ctx, query, args...)
}

// Prepare creates a prepared statement for use within a transaction.
//
// The returned statement operates within the transaction and can no longer be used once the
// transaction has been committed or rolled back.
//
// To use an existing prepared statement on this transaction, see DBTx.Stmt.
func (tx *DBTx) Prepare(query string) (*DBStmt, error) {
	stmt, err := tx.Tx.Preparex(query)
	return &DBStmt{stmt, tx.logger, query}, err
}

// PrepareContext creates a prepared statement for use within a transaction.
//
// The returned statement operates within the transaction and will be closed when the transaction
// has been committed or rolled back.
//
// To use an existing prepared statement on this transaction, see DBTx.Stmt.
//
// The provided context will be used for the preparation of the context, not for the execution of
// the returned statement. The returned statement will run in the transaction context.
func (tx *DBTx) PrepareContext(ctx context.Context, query string) (*DBStmt, error) {
	stmt, err := tx.Tx.PreparexContext(ctx, query)
	return &DBStmt{stmt, tx.logger, query}, err
}

// Query executes a query that returns rows, typically a SELECT.
func (tx *DBTx) Query(query string, args ...interface{}) (*DBRows, error) {
	tx.logger.Infof(formatDBQuery(query), args...)

	rows, err := tx.Tx.Queryx(query, args...)
	return &DBRows{rows}, err
}

// QueryContext executes a query that returns rows, typically a SELECT.
func (tx *DBTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*DBRows, error) {
	tx.logger.Infof(formatDBQuery(query), args...)

	rows, err := tx.Tx.QueryxContext(ctx, query, args...)
	return &DBRows{rows}, err
}

// QueryRow executes a query that is expected to return at most one row. QueryRow always returns
// a non-nil value. Errors are deferred until Row's Scan method is called. If the query selects no
// rows, the *DBRow's Scan will return ErrNoRows. Otherwise, the *DBRow's Scan scans the first
// selected row and discards the rest.
func (tx *DBTx) QueryRow(query string, args ...interface{}) *DBRow {
	tx.logger.Infof(formatDBQuery(query), args...)
	return &DBRow{tx.Tx.QueryRowx(query, args...)}
}

// QueryRowContext executes a query that is expected to return at most one row. QueryRowContext
// always returns a non-nil value. Errors are deferred until Row's Scan method is called. If the
// query selects no rows, the *DBRow's Scan will return ErrNoRows. Otherwise, the *DBRow's Scan scans
// the first selected row and discards the rest.
func (tx *DBTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *DBRow {
	tx.logger.Infof(formatDBQuery(query), args...)
	return &DBRow{tx.Tx.QueryRowxContext(ctx, query, args...)}
}

// Rollback aborts the transaction.
func (tx *DBTx) Rollback() error {
	tx.logger.Info(formatDBQuery("ROLLBACK;"))
	return tx.Tx.Rollback()
}

// Stmt returns a transaction-specific prepared statement from an existing statement.
//
// Example:
//  updateMoney, err := db.Prepare("UPDATE balance SET money=money+? WHERE id=?")
//  ...
//  tx, err := db.Begin()
//  ...
//  res, err := tx.Stmt(updateMoney).Exec(123.45, 98293203)
//
// The returned statement operates within the transaction and will be closed when the transaction
// has been committed or rolled back.
func (tx *DBTx) Stmt(stmt *DBStmt) *DBStmt {
	return &DBStmt{tx.Tx.Stmtx(stmt.Stmt), tx.logger, stmt.query}
}

// StmtContext returns a transaction-specific prepared statement from an existing statement.
//
// Example:
//  updateMoney, err := db.Prepare("UPDATE balance SET money=money+? WHERE id=?")
//  ...
//  tx, err := db.Begin()
//  ...
//  res, err := tx.StmtContext(ctx, updateMoney).Exec(123.45, 98293203)
//
// The provided context is used for the preparation of the statement, not for the execution of the
// statement.
//
// The returned statement operates within the transaction and will be closed when the transaction
// has been committed or rolled back.
func (tx *DBTx) StmtContext(ctx context.Context, stmt *DBStmt) *DBStmt {
	return &DBStmt{tx.Tx.StmtxContext(ctx, stmt.Stmt), tx.logger, stmt.query}
}
