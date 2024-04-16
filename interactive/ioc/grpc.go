package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	grpc2 "webook/interactive/grpc"
	"webook/pkg/grpcx"
)

func NewGrpcxServer(intrSvc *grpc2.InteractiveServiceServer) *grpcx.Server {
	server := grpc.NewServer()
	intrSvc.Register(server)
	return &grpcx.Server{
		Server: server,
		Addr:   viper.GetString("grpc.server.addr"),
	}
}
