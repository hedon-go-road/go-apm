package dogapm

import (
	"regexp"

	"github.com/hedon-go-road/go-apm/dogapm/internal"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

const (
	MetricTypeHTTP = "http"
	MetricTypeGRPC = "grpc"

	LibraryTypeMySQL = "mysql"
	LibraryTypeRedis = "redis"
)

func init() {
	MetricsReg.MustRegister(serverHandleHistogram, serverHandleCounter, clientHandleCounter, clientHandleHistogram, libraryCounter)
	MetricsReg.MustRegister(
		collectors.NewGoCollector(
			collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{
				Matcher: regexp.MustCompile("/.*"),
			}),
		),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
}

type customMetricRegistry struct {
	*prometheus.Registry
	customLabels []*io_prometheus_client.LabelPair
}

var (
	MetricsReg = newCustomMetricRegistry(map[string]string{
		"host": internal.BuildInfo.Hostname(),
		"app":  internal.BuildInfo.AppName(),
	})
)

var (
	serverHandleHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "server_handle_seconds",
		Help: "The duration of the server handle",
	}, []string{"type", "method", "status", "peer", "peer_host"})

	serverHandleCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "server_handle_total",
		Help: "The total number of server handle",
	}, []string{"type", "method", "peer", "peer_host"})

	clientHandleCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "client_handle_total",
		Help: "The total number of client handle",
	}, []string{"type", "method", "server"})

	clientHandleHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "client_handle_seconds",
		Help: "The duration of the client handle",
	}, []string{"type", "method", "server"})

	libraryCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "lib_handle_total",
		Help: "The total number of third party library handle",
	}, []string{"type", "method", "name", "server"})
)

func newCustomMetricRegistry(labels map[string]string) *customMetricRegistry {
	c := &customMetricRegistry{
		Registry: prometheus.NewRegistry(),
	}

	for k, v := range labels {
		kCp := k
		vCp := v
		c.customLabels = append(c.customLabels, &io_prometheus_client.LabelPair{
			Name:  &kCp,
			Value: &vCp,
		})
	}

	return c
}

func (c *customMetricRegistry) Gather() ([]*io_prometheus_client.MetricFamily, error) {
	metricFamilies, err := c.Registry.Gather()
	for _, mf := range metricFamilies {
		metrics := mf.Metric
		for _, metric := range metrics {
			metric.Label = append(metric.Label, c.customLabels...)
		}
	}
	return metricFamilies, err
}
