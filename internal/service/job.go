package service

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/pkg/logger"
)

type CronJobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
}

type cronJobService struct {
	repo            repository.CronJobRepository
	l               logger.LoggerV1
	refreshInterval time.Duration
}

func (c *cronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	job, err := c.repo.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	ticker := time.NewTimer(c.refreshInterval)
	go func() {
		for range ticker.C {
			c.refresh(job.Id)
		}
	}()
	job.CancelFunc = func() {
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := c.repo.Release(ctx, job.Id)
		if er != nil {
			c.l.Warn("释放job失败", logger.Error(er), logger.Int64("jid", job.Id))
		}
	}
	return job, nil
}

func (c *cronJobService) ResetNextTime(ctx context.Context, j domain.Job) error {
	nextTime := j.NextTime()
	return c.repo.ResetNextTime(ctx, j, nextTime)
}

func (c *cronJobService) refresh(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.repo.UpdateUtime(ctx, id)
	if err != nil {
		c.l.Warn("续约失败", logger.Error(err), logger.Int64("jid", id))
	}
}
