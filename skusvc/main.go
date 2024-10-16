package main

import (
	"github.com/hedon-go-road/go-apm/dogapm"
	"github.com/hedon-go-road/go-apm/protos"
	"github.com/hedon-go-road/go-apm/skusvc/api"
)

func main() {
	dogapm.Infra.Init(
		dogapm.WithMySQL("root:root@tcp(apm-mysql:3306)/skusvc?charset=utf8mb4&parseTime=True&loc=Local"),
		dogapm.WithEnableAPM("apm-otel-collector:4317"),
		dogapm.WithMetric(),
	)

	dogapm.NewHTTPServer(":30013")
	gs := dogapm.NewGrpcServer(":30003")
	protos.RegisterSkuServiceServer(gs, &api.SkuService{})

	dogapm.EndPoint.Start()
}
