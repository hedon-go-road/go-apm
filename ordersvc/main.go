package main

import (
	"net/http"

	"github.com/hedon-go-road/go-apm/dogapm"
	"github.com/hedon-go-road/go-apm/ordersvc/api"
)

func main() {
	dogapm.Infra.Init(
		dogapm.WithMySQL("root:root@tcp(127.0.0.1:23306)/ordersvc?charset=utf8mb4&parseTime=True&loc=Local"),
	)

	hs := dogapm.NewHTTPServer(":10001")
	hs.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})
	hs.HandleFunc("/order/add", api.Order.Add)

	dogapm.EndPoint.Start()
}
