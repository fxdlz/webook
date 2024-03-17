package service

import (
	"context"
	"errors"
	"time"
	"webook/internal/domain"
	"webook/internal/events/article"
	"webook/internal/repository"
	"webook/pkg/logger"
)

//go:generate mockgen -source=./article.go -package=svcmocks -destination=./mocks/article.mock.go ArticleService
type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, id int64, uid int64) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, uid, id int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	producer article.Producer
	//v1
	readerRepo repository.ArticleReaderRepository
	authorRepo repository.ArticleAuthorRepository
	l          logger.LoggerV1
}

func (a *articleService) ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error) {
	return a.repo.ListPub(ctx, start, offset, limit)
}

func (a *articleService) GetPubById(ctx context.Context, uid, id int64) (domain.Article, error) {
	res, err := a.repo.GetPubById(ctx, id)
	go func() {
		if err == nil {
			//发送消息
			er := a.producer.ProduceReadEvent(article.ReadEvent{
				Aid: id,
				Uid: uid,
			})
			if er != nil {
				a.l.Error("发送 ReadEvent 失败",
					logger.Error(er),
					logger.Int64("uid", uid),
					logger.Int64("aid", id))
			}
		}
	}()
	return res, err
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.GetByAuthor(ctx, uid, offset, limit)
}

func (a *articleService) Withdraw(ctx context.Context, id int64, uid int64) error {
	return a.repo.SyncStatus(ctx, id, uid, domain.ArticleStatusPrivate)
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, art)
}

func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)

	if art.Id > 0 {
		err = a.authorRepo.Update(ctx, art)
	} else {
		id, err = a.authorRepo.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id

	for i := 0; i < 3; i++ {
		err = a.readerRepo.Save(ctx, art)
		if err != nil {
			a.l.Error("保存到制作库成功但是到线上库失败", logger.Int64("aid", art.Id), logger.Error(err))
		} else {
			return id, nil
		}
	}
	a.l.Error("保存到制作库成功但是到线上库失败，重试耗尽", logger.Int64("aid", art.Id), logger.Error(err))
	return id, errors.New("保存到线上库失败，重试次数耗尽")
}

func NewArticleServiceV1(authorRepo repository.ArticleAuthorRepository, readerRepo repository.ArticleReaderRepository, l logger.LoggerV1) *articleService {
	return &articleService{
		readerRepo: readerRepo,
		authorRepo: authorRepo,
		l:          l,
	}
}

func NewArticleService(repo repository.ArticleRepository, producer article.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
	}
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	} else {
		return a.repo.Create(ctx, art)
	}

}
