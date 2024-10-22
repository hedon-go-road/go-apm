package main

import (
	"net/http"

	"github.com/hedon-go-road/go-apm/dogapm"
	"github.com/hedon-go-road/go-apm/ordersvc/api"
	"github.com/hedon-go-road/go-apm/ordersvc/grpcclient"
	"github.com/hedon-go-road/go-apm/ordersvc/metric"
	"github.com/hedon-go-road/go-apm/protos"
)

func main() {
	// init infra
	dogapm.Infra.Init(
		dogapm.WithMySQL("root:root@tcp(apm-mysql:3306)/ordersvc?charset=utf8mb4&parseTime=True&loc=Local"),
		dogapm.WithEnableAPM("apm-otel-collector:4317", "/logs", 10),
		dogapm.WithMetric(metric.All()...),
		dogapm.WithAutoPProf(&dogapm.AutoPProfOpt{
			EnableCPU:       true,
			EnableMem:       true,
			EnableGoroutine: true,
		}),
	)

	// init grpc clients
	grpcclient.UserClient = protos.NewUserServiceClient(dogapm.NewGrpcClient("apm-ubuntu:30002", "usrsvc"))
	grpcclient.SkuClient = protos.NewSkuServiceClient(dogapm.NewGrpcClient("apm-ubuntu:30003", "skusvc"))

	// init http server
	hs := dogapm.NewHTTPServer(":30001")
	hs.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})
	hs.HandleFunc("/order/add", api.Order.Add)

	// start all servers
	dogapm.EndPoint.Start()
}
