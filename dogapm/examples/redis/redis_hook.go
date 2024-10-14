package main

import (
	"context"
	"fmt"

	"github.com/hedon-go-road/go-apm/dogapm"
)

func main() {
	dogapm.Infra.Init(
		dogapm.WithRedis("127.0.0.1:26379"),
		dogapm.WithEnableAPM("127.0.0.1:4317"),
	)
	defer dogapm.EndPoint.Close()

	res, err := dogapm.Infra.RDB.Get(context.TODO(), "haha").Result()
	fmt.Println(res, err)
}
