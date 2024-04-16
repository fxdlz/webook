package cache

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
	"webook/interactive/domain"
)

//go:embed lua/incr_cnt.lua
var luaCnt string

const fieldReadCnt = "read_cnt"
const fieldLikeCnt = "like_cnt"
const fieldCollectCnt = "collect_cnt"

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, id int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, id int64, res domain.Interactive) error
	GetLikeTopN(ctx context.Context, biz string, num int64) ([]domain.InteractiveArticle, error)
	SetLikeTopN(ctx context.Context, biz string, num int64, data []domain.InteractiveArticle) error
}

type InteractiveRedisCache struct {
	client redis.Cmdable
}

func (i *InteractiveRedisCache) SetLikeTopN(ctx context.Context, biz string, num int64, data []domain.InteractiveArticle) error {
	arts, err := json.Marshal(data)
	if err != nil {
		return err
	}
	key := i.LikeTopNKey(biz, num)
	return i.client.Set(ctx, key, arts, 0).Err()
}

func (i *InteractiveRedisCache) GetLikeTopN(ctx context.Context, biz string, num int64) ([]domain.InteractiveArticle, error) {
	key := i.LikeTopNKey(biz, num)
	val, err := i.client.Get(ctx, key).Bytes()
	if err != nil {
		return []domain.InteractiveArticle{}, err
	}
	var res []domain.InteractiveArticle
	err = json.Unmarshal(val, &res)
	return res, err
}

func (i *InteractiveRedisCache) LikeTopNKey(biz string, num int64) string {
	return fmt.Sprintf("interactive:like:top:%s:%d", biz, num)
}

func (i *InteractiveRedisCache) IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error {
	key := i.key(biz, id)
	return i.client.Eval(ctx, luaCnt, []string{key}, fieldLikeCnt, 1).Err()
}

func (i *InteractiveRedisCache) DecrLikeCntIfPresent(ctx context.Context, biz string, id int64) error {
	key := i.key(biz, id)
	return i.client.Eval(ctx, luaCnt, []string{key}, fieldLikeCnt, -1).Err()
}

func NewInteractiveRedisCache(client redis.Cmdable) InteractiveCache {
	return &InteractiveRedisCache{
		client: client,
	}
}

func (i *InteractiveRedisCache) Set(ctx context.Context, biz string, id int64, res domain.Interactive) error {
	key := i.key(biz, id)
	err := i.client.HSet(ctx, key,
		fieldReadCnt, res.ReadCnt,
		fieldLikeCnt, res.LikeCnt,
		fieldCollectCnt, res.CollectCnt).Err()
	if err != nil {
		return err
	}
	return i.client.Expire(ctx, key, time.Minute*15).Err()
}

func (i *InteractiveRedisCache) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	key := i.key(biz, id)
	res, err := i.client.HGetAll(ctx, key).Result()
	if err != nil {
		return domain.Interactive{}, err
	}
	if len(res) == 0 {
		return domain.Interactive{}, ErrKeyNotExist
	}
	var intr domain.Interactive
	intr.CollectCnt, _ = strconv.ParseInt(res[fieldCollectCnt], 10, 64)
	intr.LikeCnt, _ = strconv.ParseInt(res[fieldLikeCnt], 10, 64)
	intr.ReadCnt, _ = strconv.ParseInt(res[fieldReadCnt], 10, 64)
	return intr, nil
}

func (i *InteractiveRedisCache) IncrCollectCntIfPresent(ctx context.Context, biz string, id int64) error {
	key := i.key(biz, id)
	return i.client.Eval(ctx, luaCnt, []string{key}, fieldCollectCnt, 1).Err()
}

func (i *InteractiveRedisCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	key := i.key(biz, bizId)
	err := i.client.Eval(ctx, luaCnt, []string{key}, fieldReadCnt, 1).Err()
	return err
}

func (i *InteractiveRedisCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}
