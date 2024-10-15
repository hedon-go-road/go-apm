package dogapm

import (
	"context"
	"database/sql/driver"
	"errors"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Hooks struct {
	Before  func(ctx context.Context, query string, args ...any) (context.Context, error)
	After   func(ctx context.Context, query string, args ...any) (context.Context, error)
	OnError func(ctx context.Context, err error, query string, args ...any) error
}

type Driver struct {
	driver.Driver
	hooks Hooks
}

type Conn struct {
	driver.Conn
	hooks Hooks
}

func (drv *Driver) Open(name string) (driver.Conn, error) {
	conn, err := drv.Driver.Open(name)
	if err != nil {
		return conn, err
	}

	wrapped := &Conn{conn, drv.hooks}
	return wrapped, nil
}

// nolint:dupl
func (conn *Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	var err error

	list := namedToAny(args)

	if ctx, err = conn.hooks.Before(ctx, query, list...); err != nil {
		return nil, err
	}

	results, err := conn.execContext(ctx, query, args)
	if err != nil {
		return results, conn.hooks.OnError(ctx, err, query, list...)
	}

	if _, err := conn.hooks.After(ctx, query, list...); err != nil {
		return nil, err
	}

	return results, err
}

func (conn *Conn) execContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	switch c := conn.Conn.(type) {
	case driver.ExecerContext:
		return c.ExecContext(ctx, query, args)
	default:
		return nil, errors.New("ExecerContext created for a non Execer driver.Conn")
	}
}

func (conn *Conn) queryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if c, ok := conn.Conn.(driver.QueryerContext); ok {
		return c.QueryContext(ctx, query, args)
	}

	// This should not happen
	return nil, errors.New("conn.Conn not implement driver.QueryerContext")
}

// nolint:dupl
func (conn *Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	var err error

	list := namedToAny(args)

	if ctx, err = conn.hooks.Before(ctx, query, list...); err != nil {
		return nil, err
	}

	results, err := conn.queryContext(ctx, query, args)
	if err != nil {
		return results, conn.hooks.OnError(ctx, err, query, list...)
	}

	if _, err := conn.hooks.After(ctx, query, list...); err != nil {
		return nil, err
	}

	return results, err
}

func namedToAny(args []driver.NamedValue) []any {
	list := make([]any, len(args))
	for i, a := range args {
		list[i] = a.Value
	}
	return list
}

func (conn *Conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	var (
		stmt driver.Stmt
		err  error
	)
	if c, ok := conn.Conn.(driver.ConnPrepareContext); ok {
		stmt, err = c.PrepareContext(ctx, query)
	}
	if err != nil {
		return stmt, err
	}
	return &Stmt{stmt, conn.hooks, query}, nil
}

type Stmt struct {
	driver.Stmt
	hooks Hooks
	query string
}

func (stmt *Stmt) execContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	if s, ok := stmt.Stmt.(driver.StmtExecContext); ok {
		return s.ExecContext(ctx, args)
	}

	return nil, errors.New("stmt.Stmt not implement driver.StmtExecContext")
}

// nolint:dupl
func (stmt *Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	var err error

	list := namedToAny(args)

	if ctx, err = stmt.hooks.Before(ctx, stmt.query, list...); err != nil {
		return nil, err
	}

	results, err := stmt.execContext(ctx, args)
	if err != nil {
		return results, stmt.hooks.OnError(ctx, err, stmt.query, list...)
	}

	if _, err := stmt.hooks.After(ctx, stmt.query, list...); err != nil {
		return nil, err
	}

	return results, err
}

func (stmt *Stmt) queryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if s, ok := stmt.Stmt.(driver.StmtQueryContext); ok {
		return s.QueryContext(ctx, args)
	}

	return nil, errors.New("stmt.Stmt not implement driver.StmtQueryContext")
}

// nolint:dupl
func (stmt *Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	var err error

	list := namedToAny(args)

	if ctx, err = stmt.hooks.Before(ctx, stmt.query, list...); err != nil {
		return nil, err
	}

	rows, err := stmt.queryContext(ctx, args)
	if err != nil {
		return rows, stmt.hooks.OnError(ctx, err, stmt.query, list...)
	}

	if _, err := stmt.hooks.After(ctx, stmt.query, list...); err != nil {
		return nil, err
	}

	return rows, err
}

type DriverTx struct {
	driver.Tx
	start time.Time
	ctx   context.Context
}

// BeginTx starts and returns a new transaction.
// If the context is canceled by the user the sql package will
// call Tx.Rollback before discarding and closing the connection.
//
// This must check opts.Isolation to determine if there is a set
// isolation level. If the driver does not support a non-default
// level and one is set or if there is a non-default isolation level
// that is not supported, an error must be returned.
//
// This must also check opts.ReadOnly to determine if the read-only
// value is true to either set the read-only transaction property if supported
// or return an error if it is not supported.
func (conn *Conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	tx, err := conn.beginTx(ctx, opts)
	if err != nil {
		return tx, err
	}

	return &DriverTx{tx, time.Now(), ctx}, nil
}

func (conn *Conn) beginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if c, ok := conn.Conn.(driver.ConnBeginTx); ok {
		return c.BeginTx(ctx, opts)
	}

	return nil, errors.New("conn.Conn not implement driver.ConnBeginTx")
}

func (tx *DriverTx) Commit() error {
	err := tx.Tx.Commit()
	elapsed := time.Since(tx.start)
	if elapsed >= longTxThreshold {
		if span := trace.SpanFromContext(tx.ctx); span != nil {
			span.SetAttributes(
				attribute.Bool("longtx", true),
				attribute.Int64("tx_duration_ms", elapsed.Milliseconds()),
			)
		}
	}
	return err
}

func (tx *DriverTx) Rollback() error {
	err := tx.Tx.Rollback()
	elapsed := time.Since(tx.start)
	if elapsed >= longTxThreshold {
		if span := trace.SpanFromContext(tx.ctx); span != nil {
			span.SetAttributes(
				attribute.Bool("longtx", true),
				attribute.Int64("tx_duration_ms", elapsed.Milliseconds()),
			)
		}
	}
	return err
}
