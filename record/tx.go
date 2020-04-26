package record

import (
	"context"
	"database/sql"
	"time"

	"github.com/appist/appy/support"
	"github.com/jmoiron/sqlx"
)

// Txer implements all Tx methods and is useful for mocking Tx in unit
// tests.
type Txer interface {
	Commit() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	NamedQuery(query string, arg interface{}) (*Rows, error)
	Prepare(query string) (*Stmt, error)
	PrepareContext(ctx context.Context, query string) (*Stmt, error)
	PrepareNamed(query string) (*NamedStmt, error)
	PrepareNamedContext(ctx context.Context, query string) (*NamedStmt, error)
	Query(query string, args ...interface{}) (*Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error)
	QueryRow(query string, args ...interface{}) *Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *Row
	Rollback() error
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
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
	start := time.Now()
	err := tx.Tx.Commit()
	tx.logger.Info(formatQuery("COMMIT;", time.Since(start)))

	return err
}

// Exec executes a query that doesn't return rows. For example: an INSERT and
// UPDATE.
func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := tx.Tx.Exec(query, args...)
	tx.logger.Infof(formatQuery(query, time.Since(start)), args...)

	return result, err
}

// ExecContext executes a query that doesn't return rows. For example: an INSERT
// and UPDATE.
func (tx *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := tx.Tx.ExecContext(ctx, query, args...)
	tx.logger.Infof(formatQuery(query, time.Since(start)), args...)

	return result, err
}

// Get using this transaction. Any placeholder parameters are replaced with
// supplied args. An error is returned if the result set is empty.
func (tx *Tx) Get(dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := tx.Tx.Get(dest, query, args...)
	tx.logger.Infof(formatQuery(query, time.Since(start)), args...)

	return err
}

// GetContext using this transaction. Any placeholder parameters are replaced
// with supplied args. An error is returned if the result set is empty.
func (tx *Tx) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := tx.Tx.GetContext(ctx, dest, query, args...)
	tx.logger.Infof(formatQuery(query, time.Since(start)), args...)

	return err
}

// NamedExec executes a named query within this transaction. Any named
// placeholder parameters are replaced with fields from arg.
func (tx *Tx) NamedExec(query string, arg interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := tx.Tx.NamedExec(query, arg)
	tx.logger.Info(formatQuery(query, time.Since(start), arg))

	return result, err
}

// NamedExecContext executes a named query within this transaction with the
// specified context. Any named placeholder parameters are replaced with fields
// from arg.
func (tx *Tx) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := tx.Tx.NamedExecContext(ctx, query, arg)
	tx.logger.Info(formatQuery(query, time.Since(start), arg))

	return result, err
}

// NamedQuery within this transaction. Any named placeholder parameters are
// replaced with fields from arg.
func (tx *Tx) NamedQuery(query string, arg interface{}) (*Rows, error) {
	start := time.Now()
	rows, err := tx.Tx.NamedQuery(query, arg)
	tx.logger.Info(formatQuery(query, time.Since(start), arg))

	return &Rows{rows}, err
}

// Prepare creates a prepared statement for use within a transaction.
//
// The returned statement operates within the transaction and can no longer be
// used once the transaction has been committed or rolled back.
//
// To use an existing prepared statement on this transaction, see Tx.Stmt.
func (tx *Tx) Prepare(query string) (*Stmt, error) {
	start := time.Now()
	stmt, err := tx.Tx.Preparex(query)
	tx.logger.Infof(formatQuery(query, time.Since(start)))

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

// PrepareNamed returns an NamedStmt.
func (tx *Tx) PrepareNamed(query string) (*NamedStmt, error) {
	start := time.Now()
	stmt, err := tx.Tx.PrepareNamed(query)
	tx.logger.Infof(formatQuery(query, time.Since(start)))

	return &NamedStmt{stmt}, err
}

// PrepareNamedContext returns an NamedStmt.
func (tx *Tx) PrepareNamedContext(ctx context.Context, query string) (*NamedStmt, error) {
	start := time.Now()
	stmt, err := tx.Tx.PrepareNamedContext(ctx, query)
	tx.logger.Infof(formatQuery(query, time.Since(start)))

	return &NamedStmt{stmt}, err
}

// Query executes a query that returns rows, typically a SELECT.
func (tx *Tx) Query(query string, args ...interface{}) (*Rows, error) {
	start := time.Now()
	rows, err := tx.Tx.Queryx(query, args...)
	tx.logger.Infof(formatQuery(query, time.Since(start)), args...)

	return &Rows{rows}, err
}

// QueryContext executes a query that returns rows, typically a SELECT.
func (tx *Tx) QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error) {
	start := time.Now()
	rows, err := tx.Tx.QueryxContext(ctx, query, args...)
	tx.logger.Infof(formatQuery(query, time.Since(start)), args...)

	return &Rows{rows}, err
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until
// sql.Row's Scan method is called. If the query selects no rows, the
// *sql.Row's Scan will return sql.ErrNoRows. Otherwise, the *sql.Row's
// Scan scans the first selected row and discards the rest.
func (tx *Tx) QueryRow(query string, args ...interface{}) *Row {
	start := time.Now()
	row := tx.Tx.QueryRowx(query, args...)
	tx.logger.Infof(formatQuery(query, time.Since(start)), args...)

	return &Row{row}
}

// QueryRowContext executes a query that is expected to return at most
// one row. QueryRowContext always returns a non-nil value. Errors are deferred
// until sql.Row's Scan method is called. If the query selects no rows, the
// *sql.Row's Scan will return sql.ErrNoRows. Otherwise, the *sql.Row's Scan
// scans the first selected row and discards the rest.
func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *Row {
	start := time.Now()
	row := tx.Tx.QueryRowxContext(ctx, query, args...)
	tx.logger.Infof(formatQuery(query, time.Since(start)), args...)

	return &Row{row}
}

// Rollback aborts the transaction.
func (tx *Tx) Rollback() error {
	start := time.Now()
	err := tx.Tx.Rollback()
	tx.logger.Info(formatQuery("ROLLBACK;", time.Since(start)))

	return err
}

// Select using this transaction. Any placeholder parameters are replaced with supplied args.
func (tx *Tx) Select(dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := tx.Tx.Select(dest, query, args...)
	tx.logger.Infof(formatQuery(query, time.Since(start)), args...)

	return err
}

// SelectContext using this transaction. Any placeholder parameters are replaced with supplied args.
func (tx *Tx) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := tx.Tx.SelectContext(ctx, dest, query, args...)
	tx.logger.Infof(formatQuery(query, time.Since(start)), args...)

	return err
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
