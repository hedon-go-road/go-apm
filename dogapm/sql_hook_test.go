package dogapm

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//nolint:all
func TestSqlHook(t *testing.T) {
	Infra.Init(
		WithMySQL("root:root@tcp(127.0.0.1:23306)/ordersvc"),
		WithEnableAPM("127.0.0.1:4317"),
	)
	defer EndPoint.Close()

	ctx, span := Tracer.Start(context.Background(), "test")
	defer span.End()

	// test slow sql
	_, _ = Infra.DB.QueryContext(ctx, "select *, sleep(2) from t_order limit ?;", 2)

	// test long tx
	tx, err := Infra.DB.BeginTx(ctx, nil)
	assert.Nil(t, err)
	_, _ = tx.QueryContext(ctx, "select * from t_order limit ?;", 2)
	time.Sleep(5 * time.Second)
	tx.Commit()
}
