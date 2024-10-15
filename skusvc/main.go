package main

import (
	"github.com/hedon-go-road/go-apm/dogapm"
	"github.com/hedon-go-road/go-apm/protos"
	"github.com/hedon-go-road/go-apm/skusvc/api"
)

func main() {
	dogapm.Infra.Init(
		dogapm.WithMySQL("root:root@tcp(127.0.0.1:23306)/skusvc?charset=utf8mb4&parseTime=True&loc=Local"),
		dogapm.WithEnableAPM("127.0.0.1:4317"),
	)

	gs := dogapm.NewGrpcServer(":30003")
	protos.RegisterSkuServiceServer(gs, &api.SkuService{})

	dogapm.EndPoint.Start()
}
