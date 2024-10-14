package main

import (
	"context"
	"fmt"

	"github.com/hedon-go-road/go-apm/dogapm"
)

func main() {
	dogapm.Infra.Init(
		dogapm.WithMySQL("root:root@tcp(127.0.0.1:23306)/ordersvc"),
		dogapm.WithEnableAPM("127.0.0.1:4317"),
	)
	defer dogapm.EndPoint.Close()

	//nolint:all
	ctx, span := dogapm.Tracer.Start(context.Background(), "sql_hook_example")
	defer span.End()
	res, err := dogapm.Infra.DB.QueryContext(ctx, "select *, sleep(1) from t_order limit ?;", 1)
	if err != nil {
		fmt.Println("query err: ", err)
	}
	if res.Err() != nil {
		fmt.Println("res err: ", res.Err())
	}
	fmt.Println("res: ", res)
	res.Close()
}
