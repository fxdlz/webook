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
	_ "webook/pkg/grpcx/balancer/wrr"
)

type BalancerTestSuite struct {
	suite.Suite
	cli *etcdv3.Client
}

func (s *BalancerTestSuite) SetupSuite() {
	cli, err := etcdv3.NewFromURL("localhost:12379")
	require.NoError(s.T(), err)
	s.cli = cli
}

func (s *BalancerTestSuite) TestFailoverClient() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(s.T(), err)
	conn, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(`
{
  "loadBalancingConfig": [{"round_robin": {}}],
  "methodConfig":  [
    {
      "name": [{"service":  "UserService"}],
      "retryPolicy": {
        "maxAttempts": 4,
        "initialBackoff": "0.01s",
        "maxBackoff": "0.1s",
        "backoffMultiplier": 2.0,
        "retryableStatusCodes": ["UNAVAILABLE"]
      }
    }
  ]
}
`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()
	uc := NewUserServiceClient(conn)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := uc.GetById(ctx, &GetByIdRequest{})
		cancel()
		require.NoError(t, err)
		t.Log(resp.User)
	}
}

func (s *BalancerTestSuite) TestClient() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(s.T(), err)
	conn, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(`
{
    "loadBalancingConfig": [
        {
            "custom_weighted_round_robin": {}
        }
    ]
}
`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()
	uc := NewUserServiceClient(conn)
	for i := 0; i < 20; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, _ := uc.GetById(ctx, &GetByIdRequest{})
		cancel()
		//require.NoError(t, err)
		t.Log(resp)
	}
}

func (s *BalancerTestSuite) TestServer() {
	go func() {
		s.startServer(":8091", 10, &Server{
			Name: "8091",
		})
	}()
	go func() {
		s.startServer(":8092", 30, &Server{
			Name: "8092",
		})
	}()
	s.startServer(":8090", 20, &FailoverServer{
		Name: "8090",
	})
}

func (s *BalancerTestSuite) startServer(adr string, weight int, svc UserServiceServer) {
	t := s.T()
	em, err := endpoints.NewManager(s.cli, "service/user")
	require.NoError(t, err)
	addr := "127.0.0.1" + adr
	key := "service/user/" + addr
	lis, err := net.Listen("tcp", adr)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var ttl int64 = 5
	leaseResp, err := s.cli.Grant(ctx, ttl)
	require.NoError(t, err)
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
		Metadata: map[string]int{
			"weight": weight,
		},
	}, etcdv3.WithLease(leaseResp.ID))
	require.NoError(t, err)

	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		ch, err1 := s.cli.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(t, err1)
		for _ = range ch {
			//t.Log(resp.String())
		}
	}()

	gs := grpc.NewServer()
	RegisterUserServiceServer(gs, svc)
	gs.Serve(lis)
	kaCancel()
	err = em.DeleteEndpoint(ctx, key)
	if err != nil {
		t.Log(err)
	}
	gs.GracefulStop()
	s.cli.Close()
}

func TestBalancer(t *testing.T) {
	suite.Run(t, new(BalancerTestSuite))
}
