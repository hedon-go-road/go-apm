package dogapm

import (
	"context"
	"fmt"
	"testing"
)

func TestRedisHook(t *testing.T) {
	Infra.Init(
		WithRedis("127.0.0.1:26379"),
		WithEnableAPM("127.0.0.1:4317"),
	)
	defer EndPoint.Close()

	res, err := Infra.RDB.Get(context.TODO(), "haha").Result()
	fmt.Println(res, err)
}
