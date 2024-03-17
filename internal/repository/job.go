package repository

import (
	"context"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

type CronJobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job, t time.Time) error
	UpdateUtime(ctx context.Context, id int64) error
	Release(ctx context.Context, id int64) error
}

type PreemptJobRepository struct {
	dao dao.JobDAO
}

func (p *PreemptJobRepository) Release(ctx context.Context, id int64) error {
	return p.dao.Release(ctx, id)
}

func (p *PreemptJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	job, err := p.dao.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}

	return domain.Job{
		Id:         job.Id,
		Name:       job.Name,
		Executor:   job.Executor,
		Expression: job.Expression,
	}, nil
}

func (p *PreemptJobRepository) ResetNextTime(ctx context.Context, j domain.Job, t time.Time) error {
	return p.dao.UpdateNextTime(ctx, j.Id, t)
}

func (p *PreemptJobRepository) UpdateUtime(ctx context.Context, id int64) error {
	return p.dao.UpdateUtime(ctx, id)
}
