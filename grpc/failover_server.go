package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FailoverServer struct {
	Name string
	UnimplementedUserServiceServer
}

func (s *FailoverServer) GetById(ctx context.Context, request *GetByIdRequest) (*GetByIdResponse, error) {
	fmt.Println("failover")
	return nil, status.Errorf(codes.Unavailable, "failover")
}
