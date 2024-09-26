package dogapm

import (
	"context"
	"net"

	"google.golang.org/grpc"
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
		if err := s.Serve(lis); err != nil {
			panic("GRPC Server failed to serve: " + err.Error())
		}
	}()
}

func (s *GrpcServer) Close() {
	s.Server.GracefulStop()
}

func unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}
}
