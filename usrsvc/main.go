package main

import (
	"github.com/hedon-go-road/go-apm/dogapm"
	"github.com/hedon-go-road/go-apm/protos"
	"github.com/hedon-go-road/go-apm/usrsvc/api"
)

func main() {
	dogapm.Infra.Init(
		dogapm.WithMySQL("root:root@tcp(127.0.0.1:23306)/usrsvc?charset=utf8mb4&parseTime=True&loc=Local"),
		dogapm.WithRedis("127.0.0.1:26379"),
	)

	hs := dogapm.NewGrpcServer(":10002")
	protos.RegisterUserServiceServer(hs.Server, &api.User{})

	dogapm.EndPoint.Start()
}
