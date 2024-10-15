package dogapm

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	grpcServerTracerName = "dogapm/grpcServer"
)

type GrpcServer struct {
	addr string
	*grpc.Server
}

func NewGrpcServer(addr string) *GrpcServer {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(unaryServerInterceptor()),
	)
	s := &GrpcServer{addr: addr, Server: server}
	globalStarters = append(globalStarters, s)
	globalClosers = append(globalClosers, s)
	return s
}

func (s *GrpcServer) Start() {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		panic("GRPC Server failed to listen: " + err.Error())
	}
	go func() {
		log.Println("starting grpc server")
		if err := s.Serve(lis); err != nil {
			panic("GRPC Server failed to serve: " + err.Error())
		}
	}()
}

func (s *GrpcServer) Close() {
	s.Server.GracefulStop()
}

func unaryServerInterceptor() grpc.UnaryServerInterceptor {
	tracer := otel.Tracer(grpcServerTracerName)

	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		// get the metadata from the incoming context or create a new one
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}

		// extract the metadata from the context
		ctx = otel.GetTextMapPropagator().Extract(ctx, &metadataSupplier{metadata: &md})

		// start the span
		ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		start := time.Now()
		defer func() {
			span.SetAttributes(attribute.String("grpc.duration_ms", fmt.Sprintf("%d", time.Since(start).Milliseconds())))
			span.End()
		}()

		// call the handler
		resp, err := handler(ctx, req)

		// set the status and error on the span
		if err != nil {
			s, _ := status.FromError(err)
			span.RecordError(err, trace.WithTimestamp(time.Now()), trace.WithStackTrace(true))
			span.SetAttributes(attribute.Bool("error", true))
			span.SetAttributes(attribute.String("grpc.status_code", s.Code().String()))
		}

		return resp, err
	}
}
