package cache

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/internal/domain"
)

type RankingCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

type RankingRedisCache struct {
	client     redis.Cmdable
	key        string
	expiration time.Duration
}

func NewRankingRedisCache(client redis.Cmdable, key string, expiration time.Duration) RankingCache {
	return &RankingRedisCache{client: client, key: key, expiration: expiration}
}

func (r *RankingRedisCache) Set(ctx context.Context, arts []domain.Article) error {
	for i, _ := range arts {
		arts[i].Content = arts[i].Abstract()
	}
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key, val, r.expiration).Err()
}

func (r *RankingRedisCache) Get(ctx context.Context) ([]domain.Article, error) {
	val, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		return nil, err
	}
	var arts []domain.Article
	err = json.Unmarshal(val, &arts)
	return arts, err
}
