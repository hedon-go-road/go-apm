package main

import (
	"net/http"

	"github.com/hedon-go-road/go-apm/dogapm"
	"github.com/hedon-go-road/go-apm/ordersvc/api"
	"github.com/hedon-go-road/go-apm/ordersvc/grpcclient"
	"github.com/hedon-go-road/go-apm/protos"
)

func main() {
	// init infra
	dogapm.Infra.Init(
		dogapm.WithMySQL("root:root@tcp(127.0.0.1:23306)/ordersvc?charset=utf8mb4&parseTime=True&loc=Local"),
	)

	// init grpc clients
	grpcclient.SkuClient = protos.NewSkuServiceClient(dogapm.NewGrpcClient(":10011"))

	// init http server
	hs := dogapm.NewHTTPServer(":10001")
	hs.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})
	hs.HandleFunc("/order/add", api.Order.Add)

	// start all servers
	dogapm.EndPoint.Start()
}
