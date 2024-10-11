package dogapm

import (
	"context"
	"database/sql"
	"sync/atomic"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestMySQLWrapper(t *testing.T) {
	beforeAtomic := atomic.Bool{}
	afterAtomic := atomic.Bool{}
	errorAtomic := atomic.Bool{}

	testDriver := &Driver{
		Driver: mysql.MySQLDriver{},
		hooks: Hooks{
			Before: func(ctx context.Context, query string, args ...any) (context.Context, error) {
				beforeAtomic.Store(true)
				return ctx, nil
			},
			After: func(ctx context.Context, query string, args ...any) error {
				afterAtomic.Store(true)
				return nil
			},
			OnError: func(ctx context.Context, err error, query string, args ...any) error {
				errorAtomic.Store(true)
				return err
			},
		},
	}

	sql.Register("test-mysql", testDriver)

	db, err := sql.Open("test-mysql", "root:root@tcp(127.0.0.1:23306)/ordersvc")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("select 1")
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, beforeAtomic.Load())
	assert.True(t, afterAtomic.Load())
	assert.False(t, errorAtomic.Load())

	_, err = db.Exec("select 1 from non_existent_table")
	assert.Error(t, err)
	assert.True(t, errorAtomic.Load())
}
