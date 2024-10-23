package dogapm

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/hedon-go-road/go-apm/dogapm/internal"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	httpTracerName = "dogapm/http"
)

type HTTPServer struct {
	mux *http.ServeMux
	*http.Server
	tracer trace.Tracer
}

func NewHTTPServer(addr string) *HTTPServer {
	mux := http.NewServeMux()
	s := &HTTPServer{
		mux: mux,
		Server: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 30 * time.Second, //nolint:mnd
		},
		tracer: otel.Tracer(httpTracerName),
	}
	s.Handle("/metrics", promhttp.HandlerFor(MetricsReg, promhttp.HandlerOpts{
		Registry: MetricsReg,
	}))
	s.Handle("/heartbeat", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	globalStarters = append(globalStarters, s)
	globalClosers = append(globalClosers, s)
	return s
}

func (s *HTTPServer) Close() {
	if s.Server != nil {
		if err := s.Server.Shutdown(context.Background()); err != nil {
			log.Println("Error stopping server:", err)
		}
	}
}

func (s *HTTPServer) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, &traceHandler{
		handler: handler,
		tracer:  s.tracer,
	})
}

func (s *HTTPServer) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.mux.Handle(pattern, &traceHandler{
		handler: http.HandlerFunc(handler),
		tracer:  s.tracer,
	})
}

func (s *HTTPServer) Start() {
	go func() {
		log.Printf("[%s][%s] starting http server on: %s\n", internal.BuildInfo.AppName(), internal.BuildInfo.Hostname(), s.Addr)
		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error starting server:", err)
		}
	}()
}

func (s *HTTPServer) Stop() {
	if err := s.Server.Shutdown(context.Background()); err != nil {
		log.Fatal("Error stopping server:", err)
	}
}

type traceHandler struct {
	handler http.Handler
	tracer  trace.Tracer
}

func (h *traceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.tracer == nil {
		h.handler.ServeHTTP(w, r)
		return
	}
	// metrics
	serverHandleCounter.WithLabelValues(MetricTypeHTTP, r.Method+"."+r.URL.Path, "", "").Inc()

	// trace
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))
	ctx, span := h.tracer.Start(ctx, "HTTP "+r.Method+" "+r.URL.Path)
	defer span.End()
	r = r.Clone(ctx)
	respWrapper := &responseWrapper{ResponseWriter: w}

	start := time.Now()
	func() {
		// panic recover
		defer func() {
			if err := recover(); err != nil {
				// TODO: log and trace
				log.Printf("Panic: %v", err)
				http.Error(respWrapper, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// handle request
		h.handler.ServeHTTP(respWrapper, r)
	}()

	if respWrapper.status == 0 {
		respWrapper.status = http.StatusOK
	}
	elapsed := time.Since(start)
	span.SetAttributes(
		attribute.KeyValue{
			Key:   "http.status_code",
			Value: attribute.IntValue(respWrapper.status),
		},
		attribute.KeyValue{
			Key:   "http.duration_ms",
			Value: attribute.Int64Value(elapsed.Milliseconds()),
		},
	)

	// metrics
	serverHandleHistogram.WithLabelValues(
		MetricTypeHTTP, r.Method+"."+r.URL.Path, strconv.Itoa(respWrapper.status), "", "",
	).Observe(elapsed.Seconds())
}

type responseWrapper struct {
	http.ResponseWriter
	status int
}

func (w *responseWrapper) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
