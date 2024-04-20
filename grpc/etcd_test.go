package grpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type EtcdTestSuite struct {
	suite.Suite
	cli *etcdv3.Client
}

func (s *EtcdTestSuite) SetupSuite() {
	cli, err := etcdv3.NewFromURL("localhost:12379")
	require.NoError(s.T(), err)
	s.cli = cli
}

func (s *EtcdTestSuite) TestClient() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(s.T(), err)
	conn, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()
	uc := NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := uc.GetById(ctx, &GetByIdRequest{})
	require.NoError(t, err)
	t.Log(resp.User)
}

func (s *EtcdTestSuite) TestServer() {
	t := s.T()
	em, err := endpoints.NewManager(s.cli, "service/user")
	require.NoError(t, err)
	addr := "127.0.0.1:8090"
	key := "service/user/" + addr
	lis, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var ttl int64 = 5
	leaseResp, err := s.cli.Grant(ctx, ttl)
	require.NoError(t, err)
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))
	require.NoError(t, err)

	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		ch, err1 := s.cli.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(t, err1)
		for resp := range ch {
			t.Log(resp.String())
		}
	}()

	gs := grpc.NewServer()
	RegisterUserServiceServer(gs, &Server{})
	gs.Serve(lis)
	kaCancel()
	err = em.DeleteEndpoint(ctx, key)
	if err != nil {
		t.Log(err)
	}
	gs.GracefulStop()
	s.cli.Close()
}

func TestEtcd(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}
