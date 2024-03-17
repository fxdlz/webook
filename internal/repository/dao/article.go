package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Article struct {
	Id      int64  `gorm:"primaryKey,autoIncrement" bson:"id"`
	Title   string `gorm:"type=varchar(4096)" bson:"title"`
	Content string `gorm:"type=BLOB" bson:"content"`
	// 我要根据创作者ID来查询
	AuthorId int64 `gorm:"index" bson:"author_id"`
	Status   uint8 `bson:"status"`
	Ctime    int64 `bson:"ctime"`
	// 更新时间
	Utime int64 `bson:"utime"`
}

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, uid int64, status uint8) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error)
}

type ArticleGORMDAO struct {
	db *gorm.DB
}

func (a *ArticleGORMDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error) {
	var res []PublishedArticle
	const ArticleStatusPublished = 2
	err := a.db.WithContext(ctx).Where("utime < ? AND status = ?", start.UnixMilli(), ArticleStatusPublished).
		Offset(offset).Limit(limit).First(&res).Error
	return res, err
}

func (a *ArticleGORMDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var art PublishedArticle
	err := a.db.WithContext(ctx).Model(&PublishedArticle{}).Where("id=?", id).First(&art).Error
	return art, err
}

func (a *ArticleGORMDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := a.db.WithContext(ctx).Model(&Article{}).Where("id=?", id).First(&art).Error
	return art, err
}

func (a *ArticleGORMDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := a.db.WithContext(ctx).Model(&Article{}).Where("author_id=?", uid).
		Offset(offset).
		Limit(limit).
		Order("utime DESC").
		Find(&arts).Error
	if err != nil {
		return nil, err
	}
	return arts, nil
}

func (a *ArticleGORMDAO) SyncStatus(ctx context.Context, id int64, uid int64, status uint8) error {
	now := time.Now().UnixMilli()
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id=? AND author_id=?", id, uid).Updates(map[string]any{
			"status": status,
			"utime":  now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return errors.New("ID不对或者创作者不对")
		}
		return tx.Model(&PublishedArticle{}).Where("id=?", id).Updates(map[string]any{
			"status": status,
			"utime":  now,
		}).Error
	})
}

func (a *ArticleGORMDAO) Sync(ctx context.Context, art Article) (int64, error) {
	tx := a.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer tx.Rollback()

	var (
		err error
		id  = art.Id
	)

	dao := NewArticleGORMDAO(tx)
	if art.Id > 0 {
		err = dao.UpdateById(ctx, art)
	} else {
		id, err = dao.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	now := time.Now().UnixMilli()
	publishArt := PublishedArticle(art)
	publishArt.Ctime = now
	publishArt.Utime = now
	err = a.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   publishArt.Title,
			"content": publishArt.Content,
			"utime":   publishArt.Utime,
			"status":  publishArt.Status,
		}),
	}).Create(&publishArt).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, nil
}

func (a *ArticleGORMDAO) SyncV1(ctx context.Context, art Article) (int64, error) {
	id := art.Id
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var (
			err error
		)

		dao := NewArticleGORMDAO(tx)
		if art.Id > 0 {
			err = dao.UpdateById(ctx, art)
		} else {
			id, err = dao.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		now := time.Now().UnixMilli()
		publishArt := PublishedArticle(art)
		publishArt.Ctime = now
		publishArt.Utime = now
		err = a.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":   publishArt.Title,
				"content": publishArt.Content,
				"utime":   publishArt.Utime,
			}),
		}).Create(&publishArt).Error
		return err
	})
	return id, err
}

func (a *ArticleGORMDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	res := a.db.WithContext(ctx).Model(&Article{}).Where("id=? AND author_id=?", art.Id, art.AuthorId).Updates(map[string]any{
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
		"utime":   now,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("ID不对或者创作者不对")
	}
	return nil
}

func NewArticleGORMDAO(db *gorm.DB) ArticleDAO {
	return &ArticleGORMDAO{
		db: db,
	}
}

func (a *ArticleGORMDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := a.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

type PublishedArticle Article
