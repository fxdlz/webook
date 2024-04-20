package grpc

import "context"

type Server struct {
	UnimplementedUserServiceServer
}

func (s *Server) GetById(ctx context.Context, request *GetByIdRequest) (*GetByIdResponse, error) {
	return &GetByIdResponse{
		User: &User{
			Phone: "123",
			Name:  "123",
		},
	}, nil
}
