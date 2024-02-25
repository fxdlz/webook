package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type ReqRetryDAO interface {
	Insert(ctx context.Context, r Reqretry) error
	Update(ctx context.Context, r Reqretry) error
	Delete(ctx context.Context, Id string) error
	FindById(ctx context.Context, Id string) (Reqretry, error)
}

type GORMReqRetryDAO struct {
	db *gorm.DB
}

func (dao *GORMReqRetryDAO) Insert(ctx context.Context, r Reqretry) error {
	now := time.Now().UnixMilli()
	r.Ctime = now
	r.Utime = now
	return dao.db.WithContext(ctx).Create(&r).Error
}

func (dao *GORMReqRetryDAO) Update(ctx context.Context, r Reqretry) error {
	r.Utime = time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Save(&r).Error
}

func (dao *GORMReqRetryDAO) Delete(ctx context.Context, Id string) error {
	return dao.db.Delete(&Reqretry{}, Id).Error
}

func (dao *GORMReqRetryDAO) FindById(ctx context.Context, Id string) (Reqretry, error) {
	var reqRetry Reqretry
	err := dao.db.WithContext(ctx).Where("Id=?", Id).First(&reqRetry).Error
	return reqRetry, err
}

func NewGORMReqRetryDAO(db *gorm.DB) *GORMReqRetryDAO {
	return &GORMReqRetryDAO{
		db: db,
	}
}

type Reqretry struct {
	Id    string `gorm:"primaryKey"`
	Req   string
	Ctime int64
	Utime int64
}
