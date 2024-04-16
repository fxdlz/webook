package repository

import (
	"context"
	"fmt"
	"github.com/gotomicro/ekit/slice"
	"gorm.io/gorm"
	"time"
	"webook/interactive/domain"
	"webook/interactive/repository/cache"
	"webook/interactive/repository/dao"
	"webook/pkg/logger"
)

var (
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error)
	BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error
	LikeTopN(ctx context.Context, biz string, num int64) ([]domain.InteractiveArticle, error)
	CronUpdateCacheLikeTopN(ctx context.Context, biz string, num int64)
	GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error)
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
	log   logger.LoggerV1
}

func (c *CachedInteractiveRepository) GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error) {
	intrArts, err := c.dao.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map(intrArts, func(idx int, src dao.Interactive) domain.Interactive {
		return c.toDomain(src)
	}), nil
}

func (c *CachedInteractiveRepository) CronUpdateCacheLikeTopN(ctx context.Context, biz string, num int64) {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				arts, er := c.dao.GetLikeTopN(ctx, biz, int(num))
				if er != nil {
					c.log.Error("获取点赞topN文章数据失败", logger.Error(er))
				}
				data := make([]domain.InteractiveArticle, len(arts))
				for i, art := range arts {
					data[i] = c.toDomainV2(art)
				}
				er = c.cache.SetLikeTopN(ctx, biz, num, data)
				if er != nil {
					c.log.Error("缓存点赞topN文章数据失败", logger.Error(er))
				} else {
					fmt.Println("缓存点赞topN文章数据成功 time:", time.Now().String())
					//c.log.Info("缓存点赞topN文章数据成功", logger.String("time", time.Now().String()))
				}
			default:
			}
		}
	}()
}

func (c *CachedInteractiveRepository) LikeTopN(ctx context.Context, biz string, num int64) ([]domain.InteractiveArticle, error) {
	intrs, err := c.cache.GetLikeTopN(ctx, biz, num)
	if err == nil {
		return intrs, nil
	}
	ies, err := c.dao.GetLikeTopN(ctx, biz, int(num))
	if err != nil {
		return []domain.InteractiveArticle{}, nil
	}
	res := make([]domain.InteractiveArticle, len(ies))
	for i, ie := range ies {
		res[i] = c.toDomainV2(ie)
	}
	go func() {
		er := c.cache.SetLikeTopN(ctx, biz, num, res)
		if er != nil {
			c.log.Error("点赞Top文章缓存失败", logger.Error(er))
		}
	}()
	return res, nil
}

func NewCachedInteractiveRepository(dao dao.InteractiveDAO, cache cache.InteractiveCache) InteractiveRepository {
	return &CachedInteractiveRepository{dao: dao, cache: cache}
}

func (c *CachedInteractiveRepository) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	intr, err := c.cache.Get(ctx, biz, id)
	if err == nil {
		return intr, nil
	}
	ie, err := c.dao.Get(ctx, biz, id)
	if err != nil {
		return domain.Interactive{}, nil
	}
	res := c.toDomain(ie)
	err = c.cache.Set(ctx, biz, id, res)
	if err != nil {
		c.log.Error("回写缓存失败", logger.Error(err), logger.String("biz", biz), logger.Int64("bizId", id))
	}
	return res, nil
}

func (c *CachedInteractiveRepository) Liked(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikedInfo(ctx, biz, id, uid)
	switch err {
	case nil:
		return true, nil
	case ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) Collected(ctx context.Context, biz string, id int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectInfo(ctx, biz, id, uid)
	switch err {
	case nil:
		return true, nil
	case ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error {
	err := c.dao.InsertCollectionBiz(ctx, dao.UserCollectionBiz{
		Cid:   cid,
		Biz:   biz,
		BizId: id,
		Uid:   uid,
	})
	if err != nil {
		return err
	}
	return c.cache.IncrCollectCntIfPresent(ctx, biz, id)
}

func (c *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	err := c.dao.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrLikeCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.DecrLikeCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	return c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error {
	err := c.dao.BatchIncrReadCnt(ctx, bizs, ids)
	if err != nil {
		return err
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		for i := 0; i < len(bizs); i++ {
			er := c.cache.IncrReadCntIfPresent(ctx, bizs[i], ids[i])
			if er != nil {
				c.log.Error("阅读数增加写缓存失败", logger.Error(er))
			}
		}
	}()
	return nil
}

func (c *CachedInteractiveRepository) toDomain(ie dao.Interactive) domain.Interactive {
	return domain.Interactive{
		BizId:      ie.BizId,
		ReadCnt:    ie.ReadCnt,
		LikeCnt:    ie.LikeCnt,
		CollectCnt: ie.CollectCnt,
	}
}

func (c *CachedInteractiveRepository) toDomainV2(ie dao.Interactive) domain.InteractiveArticle {
	return domain.InteractiveArticle{
		Id:         ie.BizId,
		ReadCnt:    ie.ReadCnt,
		LikeCnt:    ie.LikeCnt,
		CollectCnt: ie.CollectCnt,
	}
}
