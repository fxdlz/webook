package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type CachedRankingRepository struct {
	cache cache.RankingCache

	//给V1用
	redisCache *cache.RankingRedisCache
	localCache *cache.RankingLocalCache
}

func NewCachedRankingRepository(cache cache.RankingCache) RankingRepository {
	return &CachedRankingRepository{cache: cache}
}

func NewCachedRankingRepositoryV1(cache cache.RankingCache, redisCache *cache.RankingRedisCache, localCache *cache.RankingLocalCache) RankingRepository {
	return &CachedRankingRepository{cache: cache, redisCache: redisCache, localCache: localCache}
}

func (c *CachedRankingRepository) ReplaceTopNV1(ctx context.Context, arts []domain.Article) error {
	_ = c.localCache.Set(ctx, arts)
	return c.redisCache.Set(ctx, arts)
}

func (c *CachedRankingRepository) GetTopNV1(ctx context.Context) ([]domain.Article, error) {
	arts, err := c.localCache.Get(ctx)
	if err == nil {
		return arts, nil
	}
	arts, err = c.cache.Get(ctx)
	if err != nil {
		return c.localCache.ForceGet(ctx)
	}
	_ = c.localCache.Set(ctx, arts)
	return arts, nil
}

func (c *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	return c.cache.Set(ctx, arts)
}

func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return c.cache.Get(ctx)
}
