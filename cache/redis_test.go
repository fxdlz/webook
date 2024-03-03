package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"testing"
	"webook/config"
)

func TestZSet(t *testing.T) {
	var client redis.Cmdable = redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
	client.ZAdd(context.Background(), "scoreRank", redis.Z{Score: 10, Member: "yjy"})
}
