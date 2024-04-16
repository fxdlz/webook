package grpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type UserServiceTestServer struct {
	UnimplementedUserServiceServer
}

func (s *UserServiceTestServer) GetById(ctx context.Context, request *GetByIdRequest) (*GetByIdResponse, error) {
	return &GetByIdResponse{
		User: &User{
			Name: "fxlz",
		},
	}, nil
}

func TestUserServiceServer_GetById(t *testing.T) {
	lis, err := net.Listen("tcp", ":9080")
	assert.NoError(t, err)
	gs := grpc.NewServer()
	RegisterUserServiceServer(gs, &UserServiceTestServer{})
	gs.Serve(lis)
}

func TestUserServiceClient_GetById(t *testing.T) {
	conn, err := grpc.Dial(":9080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()
	uc := NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := uc.GetById(ctx, &GetByIdRequest{})
	assert.NoError(t, err)
	t.Log(resp.User)
}
