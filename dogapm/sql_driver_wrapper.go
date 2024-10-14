package dogapm

import (
	"context"
	"database/sql/driver"
	"errors"
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
	switch c := conn.Conn.(type) {
	case driver.QueryerContext:
		return c.QueryContext(ctx, query, args)
	default:
		// This should not happen
		return nil, errors.New("QueryerContext created for a non Queryer driver.Conn")
	}
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
	panic(errors.New("not implement"))
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

	panic(errors.New("not implement"))
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
