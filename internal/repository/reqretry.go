package repository

import (
	"context"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

type ReqRetryRepository interface {
	Create(ctx context.Context, r domain.ReqRetry) error
	Update(ctx context.Context, r domain.ReqRetry) error
	FindById(ctx context.Context, Id string) (domain.ReqRetry, error)
	Delete(ctx context.Context, Id string) error
}

type CodeReqRetryRepository struct {
	dao dao.ReqRetryDAO
}

func (c *CodeReqRetryRepository) toEntity(r domain.ReqRetry) dao.Reqretry {
	return dao.Reqretry{
		Id:  r.Id,
		Req: r.Req,
	}
}

func (c *CodeReqRetryRepository) toDomain(r dao.Reqretry) domain.ReqRetry {
	return domain.ReqRetry{
		Id:  r.Id,
		Req: r.Req,
	}
}

func (c *CodeReqRetryRepository) Create(ctx context.Context, r domain.ReqRetry) error {
	return c.dao.Insert(ctx, c.toEntity(r))
}

func (c *CodeReqRetryRepository) Update(ctx context.Context, r domain.ReqRetry) error {
	return c.dao.Update(ctx, c.toEntity(r))
}

func (c *CodeReqRetryRepository) FindById(ctx context.Context, Id string) (domain.ReqRetry, error) {
	r, err := c.dao.FindById(ctx, Id)
	if err != nil {
		return domain.ReqRetry{}, err
	}
	return c.toDomain(r), nil
}

func (c *CodeReqRetryRepository) Delete(ctx context.Context, Id string) error {
	return c.dao.Delete(ctx, Id)
}
