package repository

import (
	"context"
	"database/sql"
	"log"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateEmail
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	toDomain(u dao.User) domain.User
	toEntity(u domain.User) dao.User
	Update(ctx context.Context, u domain.User) error
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByWechat(ctx context.Context, openId string) (domain.User, error)
}

type CacheUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewCacheUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &CacheUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (repo *CacheUserRepository) FindByWechat(ctx context.Context, openId string) (domain.User, error) {
	u, err := repo.dao.FindByWechat(ctx, openId)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}

func (repo *CacheUserRepository) Create(ctx context.Context, u domain.User) error {
	err := repo.dao.Insert(ctx, repo.toEntity(u))
	if err == dao.ErrDuplicateEmail {
		return ErrDuplicateUser
	}
	return err
}

func (repo *CacheUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}

func (repo *CacheUserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		Nickname: u.Nickname,
		Birthday: u.Birthday,
		Profile:  u.Profile,
		WechatInfo: domain.WechatInfo{
			OpenId:  u.WechatOpenId.String,
			UnionId: u.WechatUnionId.String,
		},
	}
}

func (repo *CacheUserRepository) toEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		WechatOpenId: sql.NullString{
			String: u.WechatInfo.OpenId,
			Valid:  u.WechatInfo.OpenId != "",
		},
		WechatUnionId: sql.NullString{
			String: u.WechatInfo.UnionId,
			Valid:  u.WechatInfo.UnionId != "",
		},
		Password: u.Password,
		Nickname: u.Nickname,
		Birthday: u.Birthday,
		Profile:  u.Profile,
	}
}

func (repo *CacheUserRepository) Update(ctx context.Context, u domain.User) error {
	err := repo.dao.Update(ctx, dao.User{
		Id:       u.Id,
		Nickname: u.Nickname,
		Birthday: u.Birthday,
		Profile:  u.Profile,
	})
	return err
}

func (repo *CacheUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	du, err := repo.cache.Get(ctx, id)
	switch err {
	case nil:
		return du, nil
	case cache.ErrKeyNotExist:
		u, err := repo.dao.FindById(ctx, id)
		if err != nil {
			return domain.User{}, err
		}
		du = repo.toDomain(u)
		go func() {
			err = repo.cache.Set(ctx, du)
			if err != nil {
				log.Println(err)
			}
		}()
		return repo.toDomain(u), nil
	default:
		return domain.User{}, err

	}

}

func (repo *CacheUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := repo.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}
