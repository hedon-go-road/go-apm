package dogapm

import (
	"context"
	"fmt"
	"testing"
)

func TestSqlHook(t *testing.T) {
	Infra.Init(
		WithMySQL("root:root@tcp(127.0.0.1:23306)/ordersvc"),
		WithEnableAPM("127.0.0.1:4317"),
	)
	defer EndPoint.Close()

	//nolint:all
	ctx, span := Tracer.Start(context.Background(), "test")
	defer span.End()
	res, err := Infra.DB.QueryContext(ctx, "select *, sleep(2) from t_order limit ?;", 2)
	if err != nil {
		fmt.Println("query err: ", err)
	}
	if res.Err() != nil {
		fmt.Println("res err: ", res.Err())
	}
	fmt.Println("res: ", res)
	res.Close()
}
