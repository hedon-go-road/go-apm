package dogapm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/gops/agent"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"mosn.io/holmes"

	// import this package to fix the issue: https://github.com/open-telemetry/opentelemetry-collector/issues/10476
	// since we need to specify the version of google.golang.org/genproto, but we do not use it in the code,
	// so we need to import it to avoid deleting it by the go mod tidy command
	_ "google.golang.org/genproto/protobuf/api"

	"github.com/hedon-go-road/go-apm/dogapm/internal"
)

type infra struct {
	DB  *sql.DB
	RDB *redis.Client
}

var Infra = &infra{}
var once sync.Once

type InfraOption func(i *infra)

func WithMySQL(url string) InfraOption {
	return func(i *infra) {
		driverName := fmt.Sprintf("%s-%s", "mysql", "wrap")

		once.Do(func() {
			sql.Register(driverName, wrap(&mysql.MySQLDriver{}, url))
		})

		db, err := sql.Open(driverName, url)
		if err != nil {
			panic(err)
		}
		err = db.Ping()
		if err != nil {
			panic(err)
		}
		i.DB = db
	}
}

func WithRedis(url string) InfraOption {
	return func(i *infra) {
		client := redis.NewClient(&redis.Options{
			Addr:     url,
			DB:       0,
			Password: "",
		})
		client.AddHook(&redisHook{})
		res, err := client.Ping(context.TODO()).Result()
		if err != nil {
			panic(err)
		}
		if res != "PONG" {
			panic("redis ping failed")
		}
		i.RDB = client
	}
}

func WithEnableAPM(otelEndpoint, logPathPrefix string, maxLogCnt uint) InfraOption {
	return func(i *infra) {
		ctx := context.Background()
		// Set up a resource
		res, err := resource.New(ctx,
			resource.WithAttributes(
				semconv.ServiceName(internal.BuildInfo.AppName()),
			),
		)
		if err != nil {
			panic(err)
		}

		// Connect to otel collector
		conn, err := grpc.NewClient(otelEndpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			panic(err)
		}

		// Set up a trace exporter
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			panic(err)
		}
		bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
		tracerProvider := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithResource(res),
			sdktrace.WithSpanProcessor(bsp),
		)
		otel.SetTracerProvider(tracerProvider)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

		// Use shutdown to flush and close the tracer provider when the application exits
		globalClosers = append(globalClosers, &traceProviderComponent{tracerProvider})

		// logs
		path := filepath.Join(logPathPrefix, internal.BuildInfo.AppName()+".%Y%m%d.log")
		writer, err := rotatelogs.New(path,
			rotatelogs.WithRotationCount(maxLogCnt),
			rotatelogs.WithRotationTime(time.Hour&24*7),
		)
		if err != nil {
			panic(err)
		}
		logrus.SetOutput(writer)
	}
}

type traceProviderComponent struct {
	*sdktrace.TracerProvider
}

func (t *traceProviderComponent) Close() {
	_ = t.TracerProvider.Shutdown(context.Background())
}

func (i *infra) Init(opts ...InfraOption) {
	for _, opt := range opts {
		opt(i)
	}

	Tracer = otel.Tracer(internal.BuildInfo.AppName())
}

func WithMetric(collectors ...prometheus.Collector) InfraOption {
	return func(i *infra) {
		MetricsReg.MustRegister(collectors...)
	}
}

type AutoPProfOpt struct {
	EnableCPU       bool
	EnableMem       bool
	EnableGoroutine bool
}

type autoPProfReporter struct{}

//nolint:lll
func (a *autoPProfReporter) Report(pType, filename string, reason holmes.ReasonType, eventID string, sampleTime time.Time, pprofBytes []byte, scene holmes.Scene) error {
	Logger.Error(context.Background(), "homesGen", map[string]any{
		"reason":  reason,
		"eventID": filename,
		"pType":   pType,
	}, errors.New("auto record running state"))
	return nil
}

func WithAutoPProf(autoPProfOpts *AutoPProfOpt, opts ...holmes.Option) InfraOption {
	if err := agent.Listen(agent.Options{}); err != nil {
		panic(err)
	}

	opts = append(opts, holmes.WithProfileReporter(&autoPProfReporter{}))
	return func(i *infra) {
		h, err := holmes.New(opts...)
		if err == nil && autoPProfOpts != nil {
			if autoPProfOpts.EnableCPU {
				h.EnableCPUDump()
			}
			if autoPProfOpts.EnableMem {
				h.EnableMemDump()
			}
			if autoPProfOpts.EnableGoroutine {
				h.EnableGoroutineDump()
			}
			h.Start()
		}
	}
}
