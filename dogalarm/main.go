package main

import (
	"net/http"

	"github.com/hedon-go-road/go-apm/dogalarm/api"
	"github.com/hedon-go-road/go-apm/dogalarm/metric"
	"github.com/hedon-go-road/go-apm/dogapm"
)

func main() {
	dogapm.Infra.Init(
		dogapm.WithMySQL("root:root@tcp(apm-mysql:3306)/dogalarm?charset=utf8mb4&parseTime=True&loc=Local"),
		dogapm.WithEnableAPM("apm-otel-collector:4317", "/logs", 10),
		dogapm.WithMetric(metric.All()...),
		dogapm.WithAutoPProf(&dogapm.AutoPProfOpt{
			EnableCPU:       true,
			EnableMem:       true,
			EnableGoroutine: true,
		}),
	)

	httpServer := dogapm.NewHTTPServer(":30004")
	httpServer.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})
	httpServer.HandleFunc("/metric_webhook", api.Alarm.MetricWebHook)
	httpServer.HandleFunc("/log_webhook", api.Alarm.LogWebHook)
	dogapm.EndPoint.Start()
}
