package cache

import (
	"context"
	"errors"
	"github.com/gotomicro/ekit/syncx/atomicx"
	"time"
	"webook/internal/domain"
)

type RankingLocalCache struct {
	topN       *atomicx.Value[[]domain.Article]
	ddl        *atomicx.Value[time.Time]
	expiration time.Duration
}

func (c *RankingLocalCache) Get(ctx context.Context) ([]domain.Article, error) {
	arts := c.topN.Load()
	ddl := c.ddl.Load()
	if len(arts) == 0 || ddl.Before(time.Now()) {
		return nil, errors.New("本地缓存失效了")
	}
	return arts, nil
}

func (c *RankingLocalCache) ForceGet(ctx context.Context) ([]domain.Article, error) {
	arts := c.topN.Load()
	if len(arts) == 0 {
		return nil, errors.New("本地缓存失效了")
	}
	return arts, nil
}

func (c *RankingLocalCache) Set(ctx context.Context, arts []domain.Article) error {
	c.topN.Store(arts)
	c.ddl.Store(time.Now().Add(c.expiration))
	return nil
}
