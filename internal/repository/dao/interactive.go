package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error
	DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error
	InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error
	Get(ctx context.Context, biz string, id int64) (Interactive, error)
	GetLikedInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error)
	GetCollectInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error)
}

type GORMInteractiveDAO struct {
	db *gorm.DB
}

func (g *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{DoUpdates: clause.Assignments(map[string]interface{}{
			"status": 1,
			"utime":  now,
		})}).Create(&UserLikeBiz{
			Uid:    uid,
			BizId:  bizId,
			Biz:    biz,
			Status: 1,
			Ctime:  now,
			Utime:  now,
		}).Error
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "biz_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt": gorm.Expr("like_cnt + 1"),
				"utime":    now,
			}),
		}).Create(&Interactive{
			BizId:   bizId,
			Biz:     biz,
			LikeCnt: 1,
			Ctime:   now,
			Utime:   now,
		}).Error
	})
}

func (g *GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&UserLikeBiz{}).
			Where("uid=? AND biz_id=? AND biz=?", uid, bizId, biz).
			Updates(map[string]interface{}{
				"utime":  now,
				"status": 0,
			}).Error
		if err != nil {
			return err
		}
		return tx.Model(&Interactive{}).
			Where("biz_id=? AND biz=?", bizId, biz).
			Updates(map[string]interface{}{
				"like_cnt": gorm.Expr("like_cnt - 1"),
				"utime":    now,
			}).Error
	})
}

func NewGORMInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &GORMInteractiveDAO{db: db}
}

func (g *GORMInteractiveDAO) GetLikedInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := g.db.WithContext(ctx).Where("biz_id=? AND biz=? AND uid=? AND status=?", id, biz, uid, 1).First(&res).Error
	return res, err
}

func (g *GORMInteractiveDAO) GetCollectInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := g.db.WithContext(ctx).Where("biz_id=? AND biz=? AND uid=?", id, biz, uid).First(&res).Error
	return res, err
}

func (g *GORMInteractiveDAO) Get(ctx context.Context, biz string, id int64) (Interactive, error) {
	var res Interactive
	err := g.db.WithContext(ctx).Where("biz_id=? AND biz=?", id, biz).First(&res).Error
	return res, err
}

func (g *GORMInteractiveDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	cb.Utime = now
	cb.Ctime = now
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&cb).Error
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "biz_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"collect_cnt": gorm.Expr("collect_cnt + 1"),
				"utime":       now,
			}),
		}).Create(&Interactive{
			BizId:      cb.BizId,
			Biz:        cb.Biz,
			CollectCnt: 1,
			Ctime:      now,
			Utime:      now,
		}).Error
	})
}

func (g *GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "biz_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"read_cnt": gorm.Expr("read_cnt + 1"),
			"utime":    now,
		}),
	}).Create(&Interactive{
		BizId:   bizId,
		Biz:     biz,
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
	}).Error
}

type Interactive struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	BizId      int64  `gorm:"uniqueIndex:biz_type_id"`
	Biz        string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`
	ReadCnt    int64
	CollectCnt int64
	LikeCnt    int64
	Ctime      int64
	Utime      int64
}

type UserLikeBiz struct {
	Id     int64  `gorm:"primaryKey,autoIncrement"`
	Uid    int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId  int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz    string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	Status int
	Utime  int64
	Ctime  int64
}

type UserCollectionBiz struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	Uid   int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	Cid   int64  `gorm:"index"`
	Utime int64
	Ctime int64
}
