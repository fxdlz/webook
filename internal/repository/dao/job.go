package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateNextTime(ctx context.Context, id int64, t time.Time) error
	UpdateUtime(ctx context.Context, id int64) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func (g *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := g.db.WithContext(ctx)
	for {
		var job Job
		now := time.Now().UnixMilli()
		err := db.WithContext(ctx).
			Where("status = ? AND next_time < ?", jobStatusWaiting, now).
			First(&job).Error
		if err != nil {
			return Job{}, err
		}
		res := db.WithContext(ctx).Model(&Job{}).
			Where("id=? AND version=?", job.Id, job.Version).
			Updates(map[string]any{
				"version": job.Version + 1,
				"status":  jobStatusRunning,
				"utime":   now,
			})
		if res.Error != nil {
			return Job{}, err
		}
		if res.RowsAffected == 0 {
			continue
		}
		return job, nil
	}
}

func (g *GORMJobDAO) Release(ctx context.Context, jid int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Model(&Job{}).Where("id=?", jid).Updates(map[string]any{
		"status": jobStatusWaiting,
		"utime":  now,
	}).Error
}

func (g *GORMJobDAO) UpdateNextTime(ctx context.Context, id int64, t time.Time) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Model(&Job{}).Where("id=?", id).Updates(map[string]any{
		"next_time": t.UnixMilli(),
		"utime":     now,
	}).Error
}

func (g *GORMJobDAO) UpdateUtime(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Model(&Job{}).Where("id=?", id).Updates(map[string]any{
		"utime": now,
	}).Error
}

type Job struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"type:varchar(128);unique"`
	Executor   string
	Expression string
	Status     int
	Version    int
	NextTime   int64 `gorm:"index"`
	Ctime      int64
	// 更新时间
	Utime int64
}

const (
	// jobStatusWaiting 没人抢
	jobStatusWaiting = iota
	// jobStatusRunning 已经被人抢了
	jobStatusRunning
	// jobStatusPaused 不再需要调度了
	jobStatusPaused
)
