package dogapm

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hedon-go-road/go-apm/dogapm/internal"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
			sql.Register(driverName, wrap(&mysql.MySQLDriver{}))
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

func WithEnableAPM(otelEndpoint string) InfraOption {
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
