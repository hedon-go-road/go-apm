package dogapm

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	grpcClientTracerName = "dogapm/grpcClient"
)

type GrpcClient struct {
	*grpc.ClientConn
}

func NewGrpcClient(addr, server string) *GrpcClient {
	conn, err := grpc.NewClient(addr,
		grpc.WithUnaryInterceptor(unaryInterceptor(server)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic("GRPC Client failed to connect to server: " + err.Error())
	}
	return &GrpcClient{conn}
}

func unaryInterceptor(server string) grpc.UnaryClientInterceptor {
	tracer := otel.Tracer(grpcClientTracerName)

	return func(ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		// trace
		ctx, span := tracer.Start(ctx, method, trace.WithSpanKind(trace.SpanKindClient))
		start := time.Now()
		defer func() {
			span.SetAttributes(attribute.Int64("grpc.duration_ms", time.Since(start).Milliseconds()))
			span.End()

			// metric
			clientHandleHistogram.WithLabelValues(MetricTypeGRPC, method, server).Observe(time.Since(start).Seconds())
		}()

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}
		otel.GetTextMapPropagator().Inject(ctx, &metadataSupplier{metadata: &md})
		ctx = metadata.NewOutgoingContext(ctx, md)

		// metric
		clientHandleCounter.WithLabelValues(MetricTypeGRPC, method, server).Inc()

		// invoke the actual grpc call
		err := invoker(ctx, method, req, reply, cc, opts...)

		if err != nil {
			s, _ := status.FromError(err)
			span.RecordError(err, trace.WithTimestamp(time.Now()), trace.WithStackTrace(true))
			span.SetAttributes(attribute.Bool("error", true))
			span.SetAttributes(attribute.String("grpc.status_code", s.Code().String()))
		}
		return err
	}
}
