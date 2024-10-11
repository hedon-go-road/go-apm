package dogapm

import (
	"context"
	"database/sql/driver"
)

type Hooks struct {
	Before  func(ctx context.Context, query string, args ...any) (context.Context, error)
	After   func(ctx context.Context, query string, args ...any) error
	OnError func(ctx context.Context, err error, query string, args ...any) error
}

type Driver struct {
	driver.Driver
	hooks Hooks
}

func (d *Driver) Open(name string) (driver.Conn, error) {
	conn, err := d.Driver.Open(name)
	if err != nil {
		return nil, err
	}
	return &Conn{Conn: conn, hooks: d.hooks}, nil
}

type Conn struct {
	driver.Conn
	hooks Hooks
}

//nolint:dupl
func (c *Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	queryer, ok := c.Conn.(driver.QueryerContext)
	if !ok {
		panic("not implement driver.QueryerContext")
	}

	list := namedToAny(args)

	ctx, err := c.hooks.Before(ctx, query, list...)
	if err != nil {
		return nil, err
	}

	rows, err := queryer.QueryContext(ctx, query, args)
	if err != nil {
		return nil, c.hooks.OnError(ctx, err, query, list...)
	}

	return rows, c.hooks.After(ctx, query, list...)
}

//nolint:dupl
func (c *Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	exec, ok := c.Conn.(driver.ExecerContext)
	if !ok {
		panic("not implement driver.ExecerContext")
	}

	list := namedToAny(args)

	ctx, err := c.hooks.Before(ctx, query, list...)
	if err != nil {
		return nil, err
	}

	result, err := exec.ExecContext(ctx, query, args)
	if err != nil {
		return nil, c.hooks.OnError(ctx, err, query, list...)
	}

	return result, c.hooks.After(ctx, query, list...)
}

// PrepareContext returns a prepared statement, bound to this connection.
// context is for the preparation of the statement,
// it must not store the context within the statement itself.
func (c *Conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	prepare, ok := c.Conn.(driver.ConnPrepareContext)
	if !ok {
		panic("not implement driver.ConnPrepareContext")
	}

	stmt, err := prepare.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &Stmt{Stmt: stmt, hooks: c.hooks, query: query}, nil
}

type Stmt struct {
	driver.Stmt
	hooks Hooks
	query string
}

// QueryContext executes a query that may return rows, such as a
// SELECT.
//
// QueryContext must honor the context timeout and return when it is canceled.
//
//nolint:dupl
func (s *Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	stmt, ok := s.Stmt.(driver.StmtQueryContext)
	if !ok {
		panic("not implement driver.StmtQueryContext")
	}

	result, err := s.execWithHooks(ctx, s.query, args,
		func(ctx context.Context, args []driver.NamedValue) (interface{}, error) {
			return stmt.QueryContext(ctx, args)
		},
	)

	if err != nil {
		return nil, err
	}
	return result.(driver.Rows), nil
}

// ExecContext executes a query that doesn't return rows, such
// as an INSERT or UPDATE.
//
// ExecContext must honor the context timeout and return when it is canceled.
//
//nolint:dupl
func (s *Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	stmt, ok := s.Stmt.(driver.StmtExecContext)
	if !ok {
		panic("not implement driver.StmtExecContext")
	}

	result, err := s.execWithHooks(ctx, s.query, args,
		func(ctx context.Context, args []driver.NamedValue) (interface{}, error) {
			return stmt.ExecContext(ctx, args)
		},
	)

	if err != nil {
		return nil, err
	}
	return result.(driver.Result), nil
}

func (s *Stmt) execWithHooks(ctx context.Context, query string, args []driver.NamedValue,
	exec func(context.Context, []driver.NamedValue) (interface{}, error)) (interface{}, error) {
	list := namedToAny(args)
	ctx, err := s.hooks.Before(ctx, query, list...)
	if err != nil {
		return nil, err
	}

	result, err := exec(ctx, args)
	if err != nil {
		return nil, s.hooks.OnError(ctx, err, query, list...)
	}

	return result, s.hooks.After(ctx, query, list...)
}

func namedToAny(args []driver.NamedValue) []any {
	anyArgs := make([]any, len(args))
	for i, arg := range args {
		anyArgs[i] = arg.Value
	}
	return anyArgs
}
