package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/internal/domain"
)

type ArticleCache interface {
	SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	DelFirstPage(ctx context.Context, uid int64) error
	Get(ctx context.Context, id int64) (domain.Article, error)
	Set(ctx context.Context, res domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
	SetPub(ctx context.Context, art domain.Article) error
}

type ArticleRedisCache struct {
	client redis.Cmdable
}

func NewArticleRedisCache(client redis.Cmdable) ArticleCache {
	return &ArticleRedisCache{
		client: client,
	}
}

func (a *ArticleRedisCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	val, err := a.client.Get(ctx, a.pubKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal(val, &art)
	return art, err
}

func (a *ArticleRedisCache) SetPub(ctx context.Context, res domain.Article) error {
	art, err := json.Marshal(res)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, a.pubKey(res.Id), art, 10*time.Minute).Err()
}

func (a *ArticleRedisCache) Set(ctx context.Context, res domain.Article) error {
	art, err := json.Marshal(res)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, a.key(res.Id), art, 3*time.Second).Err()
}

func (a *ArticleRedisCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	val, err := a.client.Get(ctx, a.key(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal(val, &art)
	return art, err
}

func (a *ArticleRedisCache) SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error {
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}
	key := a.firstKey(uid)
	value, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, key, value, 10*time.Minute).Err()
}

func (a *ArticleRedisCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	key := a.firstKey(uid)
	val, err := a.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}

func (a *ArticleRedisCache) DelFirstPage(ctx context.Context, uid int64) error {
	key := a.firstKey(uid)
	err := a.client.Del(ctx, key).Err()
	return err
}

func (a *ArticleRedisCache) firstKey(uid int64) string {
	return fmt.Sprintf("article:first_page:%d", uid)
}

func (a *ArticleRedisCache) key(id int64) string {
	return fmt.Sprintf("article:detail:%d", id)
}

func (a *ArticleRedisCache) pubKey(id int64) string {
	return fmt.Sprintf("article:pub:detail:%d", id)
}
