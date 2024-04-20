package grpc

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type RedisTestSuite struct {
	suite.Suite
	cmd redis.Cmdable
}

func (s *RedisTestSuite) SetupSuite() {
	s.cmd = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func (s *RedisTestSuite) TestClient() {
	t := s.T()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	val, err := s.cmd.Get(ctx, "service-user").Result()
	require.NoError(s.T(), err)
	conn, err := grpc.Dial(val,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()
	uc := NewUserServiceClient(conn)
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := uc.GetById(ctx, &GetByIdRequest{})
	require.NoError(t, err)
	t.Log(resp.User)
}

func (s *RedisTestSuite) TestServer() {
	t := s.T()
	addr := "127.0.0.1:8090"
	key := "service-user"
	lis, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = s.cmd.Set(ctx, key, []byte(addr), time.Second*0).Err()
	require.NoError(t, err)
	gs := grpc.NewServer()
	RegisterUserServiceServer(gs, &Server{})
	gs.Serve(lis)
	gs.GracefulStop()
}

func TestRedis(t *testing.T) {
	suite.Run(t, new(RedisTestSuite))
}
