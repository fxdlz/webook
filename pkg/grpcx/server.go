package grpcx

import (
	"context"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"time"
	"webook/pkg/logger"
	"webook/pkg/netx"
)

type Server struct {
	Server   *grpc.Server
	EtcdAddr string
	Port     int
	Name     string
	L        logger.LoggerV1
	cli      *etcdv3.Client
	kaCancel func()
}

func (s *Server) Serve() error {
	addr := ":" + strconv.Itoa(s.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	err = s.register()
	if err != nil {
		return err
	}
	return s.Server.Serve(l)
}

func (s *Server) register() error {
	cli, err := etcdv3.NewFromURL(s.EtcdAddr)
	if err != nil {
		return err
	}
	s.cli = cli
	em, err := endpoints.NewManager(s.cli, "service/"+s.Name)
	if err != nil {
		return err
	}
	addr := netx.GetOutboundIP()
	key := "service/" + s.Name + "/" + addr
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var ttl int64 = 5
	leaseResp, err := s.cli.Grant(ctx, ttl)
	if err != nil {
		return err
	}
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))

	if err != nil {
		return err
	}
	kaCtx, kaCancel := context.WithCancel(context.Background())
	s.kaCancel = kaCancel
	ch, err := s.cli.KeepAlive(kaCtx, leaseResp.ID)
	if err != nil {
		return err
	}
	go func() {
		for resp := range ch {
			s.L.Debug(resp.String())
		}
	}()
	return nil
}

func (s *Server) Close() error {
	if s.kaCancel != nil {
		s.kaCancel()
	}
	if s.cli != nil {
		err := s.cli.Close()
		if err != nil {
			return err
		}
	}
	s.Server.GracefulStop()
	return nil
}
