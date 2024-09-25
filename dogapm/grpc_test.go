package dogapm

import (
	"context"
	"testing"
	"time"

	"github.com/hedon-go-road/go-apm/protos"
	"github.com/stretchr/testify/assert"
)

type helloSvc struct {
	protos.UnimplementedHelloServiceServer
}

func (s *helloSvc) SayHello(ctx context.Context, in *protos.HelloRequest) (*protos.HelloResponse, error) {
	return &protos.HelloResponse{Message: "Hello, " + in.Name}, nil
}

func TestGrpcServerAndClient(t *testing.T) {
	server := NewGrpcServer(":50051")
	protos.RegisterHelloServiceServer(server, &helloSvc{})
	server.Start()

	time.Sleep(100 * time.Millisecond)

	client := NewGrpcClient("localhost:50051")
	res, err := protos.NewHelloServiceClient(client).SayHello(context.Background(),
		&protos.HelloRequest{Name: "World"})
	assert.Nil(t, err)
	assert.Equal(t, "Hello, World", res.Message)
}
