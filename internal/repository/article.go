package repository

import (
	"context"
	"github.com/gotomicro/ekit/slice"
	"gorm.io/gorm"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, uid int64, status domain.ArticleStatus) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
}

type CacheArticleRepository struct {
	userRepo  UserRepository
	dao       dao.ArticleDAO
	cache     cache.ArticleCache
	readerDao dao.ArticleReaderDAO
	authorDao dao.ArticleAuthorDAO
	db        *gorm.DB
}

func (c *CacheArticleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.GetPub(ctx, id)
	if err == nil {
		return res, nil
	}
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = c.toDomain(dao.Article(art))
	author, err := c.userRepo.FindById(ctx, res.Author.Id)
	if err != nil {
		return domain.Article{}, err
	}
	res.Author.Name = author.Nickname

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := c.cache.SetPub(ctx, res)
		if err != nil {

		}
	}()

	return res, nil
}

func (c *CacheArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	const size = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) < size {
		err := c.cache.Set(ctx, arts[0])
		if err != nil {
			//记录日志
		}
	}
}

func (c *CacheArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := c.cache.Get(ctx, id)
	if err == nil {
		return res, err
	}
	art, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = c.toDomain(art)
	go func() {
		c.cache.Set(ctx, res)
	}()
	return res, nil
}

func (c *CacheArticleRepository) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	if offset == 0 && limit == 100 {
		res, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil && len(res) > 0 {
			return res, nil
		}
	}

	arts, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	res := slice.Map[dao.Article, domain.Article](arts, func(idx int, src dao.Article) domain.Article {
		return c.toDomain(src)
	})

	if offset == 0 && limit == 100 {
		err = c.cache.SetFirstPage(ctx, uid, res)
		if err != nil {
			//记录日志
		}
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		c.preCache(ctx, res)
	}()
	return res, nil
}

func (c *CacheArticleRepository) SyncStatus(ctx context.Context, id int64, uid int64, status domain.ArticleStatus) error {
	err := c.dao.SyncStatus(ctx, id, uid, status.ToUint8())
	if err == nil {
		c.cache.DelFirstPage(ctx, uid)
		if err != nil {
			//记录日志
		}
	}
	return err
}

func NewCacheArticleRepositoryV2(readerDao dao.ArticleReaderDAO, authorDao dao.ArticleAuthorDAO) *CacheArticleRepository {
	return &CacheArticleRepository{
		readerDao: readerDao,
		authorDao: authorDao,
	}
}

func (c *CacheArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err == nil {
		er := c.cache.DelFirstPage(ctx, art.Author.Id)
		if er != nil {
			//记录日志
		}
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		author, er := c.userRepo.FindById(ctx, art.Author.Id)
		if er != nil {
			return
		}
		art.Author = domain.Author{
			Id:   author.Id,
			Name: author.Nickname,
		}
		er = c.cache.SetPub(ctx, art)
		if er != nil {
			//记录日志
		}
	}()
	return id, err
}

func (c *CacheArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer tx.Rollback()
	authorDao := dao.NewArticleGORMAuthorDAO(tx)
	readerDao := dao.NewArticleGORMReaderDAO(tx)
	var (
		err error
		id  = art.Id
	)
	if art.Id > 0 {
		err = authorDao.Update(ctx, c.toEntity(art))
	} else {
		id, err = authorDao.Create(ctx, c.toEntity(art))
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	err = readerDao.Upsert(ctx, c.toEntity(art))
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, nil
}

func (c *CacheArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		err error
		id  = art.Id
	)
	if art.Id > 0 {
		err = c.authorDao.Update(ctx, c.toEntity(art))
	} else {
		id, err = c.authorDao.Create(ctx, c.toEntity(art))
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	err = c.readerDao.Upsert(ctx, c.toEntity(art))
	return id, err
}

func (c *CacheArticleRepository) Update(ctx context.Context, art domain.Article) error {
	err := c.dao.UpdateById(ctx, c.toEntity(art))
	if err == nil {
		c.cache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			//记录日志
		}
	}
	return err
}

func NewCacheArticleRepository(dao dao.ArticleDAO,
	userRepo UserRepository,
	cache cache.ArticleCache) ArticleRepository {
	return &CacheArticleRepository{
		dao:      dao,
		userRepo: userRepo,
		cache:    cache,
	}
}
func (c *CacheArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Insert(ctx, c.toEntity(art))
	if err == nil {
		c.cache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			//记录日志
		}
	}
	return id, err
}

func (c *CacheArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (c *CacheArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Status: domain.ArticleStatus(art.Status),
	}
}
