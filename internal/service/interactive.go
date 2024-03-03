package service

import (
	"context"
	"golang.org/x/sync/errgroup"
	"webook/internal/domain"
	"webook/internal/repository"
)

type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, id int64, uid int64) error
	CancelLike(ctx context.Context, biz string, id int64, uid int64) error
	Collect(ctx context.Context, biz string, id, cid, uid int64) error
	Get(ctx context.Context, biz string, id int64, uid int64) (domain.Interactive, error)
	LikeTopN(ctx context.Context, biz string, num int64) ([]domain.InteractiveArticle, error)
	CronUpdateCacheLikeTopN(ctx context.Context, biz string, num int64)
}

type interactiveService struct {
	repo repository.InteractiveRepository
}

func NewInteractiveService(repo repository.InteractiveRepository) InteractiveService {
	return &interactiveService{repo: repo}
}

func (i *interactiveService) CronUpdateCacheLikeTopN(ctx context.Context, biz string, num int64) {
	i.repo.CronUpdateCacheLikeTopN(ctx, biz, num)
}

func (i *interactiveService) LikeTopN(ctx context.Context, biz string, num int64) ([]domain.InteractiveArticle, error) {
	return i.repo.LikeTopN(ctx, biz, num)
}

func (i *interactiveService) Get(ctx context.Context, biz string, id int64, uid int64) (domain.Interactive, error) {
	intr, err := i.repo.Get(ctx, biz, id)
	if err != nil {
		return domain.Interactive{}, err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		var er error
		intr.Liked, er = i.repo.Liked(ctx, biz, id, uid)
		return er
	})
	eg.Go(func() error {
		var er error
		intr.Collected, er = i.repo.Collected(ctx, biz, id, uid)
		return er
	})
	return intr, eg.Wait()
}

func (i *interactiveService) Collect(ctx context.Context, biz string, id, cid, uid int64) error {
	return i.repo.AddCollectionItem(ctx, biz, id, cid, uid)
}

func (i *interactiveService) Like(ctx context.Context, biz string, id int64, uid int64) error {
	return i.repo.IncrLike(ctx, biz, id, uid)
}

func (i *interactiveService) CancelLike(ctx context.Context, biz string, id int64, uid int64) error {
	return i.repo.DecrLike(ctx, biz, id, uid)
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}
