package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hedon-go-road/go-apm/dogapm"
)

//nolint:all
func main() {
	dogapm.Infra.Init(
		dogapm.WithMySQL("root:root@tcp(127.0.0.1:23306)/ordersvc"),
		dogapm.WithEnableAPM("127.0.0.1:4317"),
	)
	defer dogapm.EndPoint.Close()

	ctx, span := dogapm.Tracer.Start(context.Background(), "sql_hook_example")
	defer span.End()

	// test slow sql
	_, _ = dogapm.Infra.DB.QueryContext(ctx, "select *, sleep(2) from t_order limit ?;", 1)

	// test long tx
	tx, err := dogapm.Infra.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault})
	if err != nil {
		fmt.Println("begin tx err: ", err)
	}
	_, _ = tx.ExecContext(ctx, "select * from t_order limit ?;", 2)
	time.Sleep(5 * time.Second)
	tx.Commit()
}
