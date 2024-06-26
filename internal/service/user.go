package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"webook/internal/domain"
	"webook/internal/repository"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("用户或密码不正确")
)

type UserService interface {
	Signup(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	Edit(ctx context.Context, u domain.User) error
	Profile(ctx context.Context, id int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
}
type CacheUserService struct {
	repo repository.UserRepository
}

func (svc *CacheUserService) FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error) {
	u, err := svc.repo.FindByWechat(ctx, wechatInfo.OpenId)
	if err != repository.ErrUserNotFound {
		return u, err
	}
	err = svc.repo.Create(ctx, domain.User{
		WechatInfo: wechatInfo,
	})
	if err != nil && err != repository.ErrDuplicateUser {
		return domain.User{}, err
	}

	return svc.repo.FindByWechat(ctx, wechatInfo.OpenId)
}

func NewCacheUserService(repo repository.UserRepository) UserService {
	return &CacheUserService{repo: repo}
}

func (svc *CacheUserService) Signup(ctx context.Context, u domain.User) error {
	encrypted, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(encrypted)
	err = svc.repo.Create(ctx, u)
	if err == repository.ErrDuplicateUser {
		return ErrDuplicateEmail
	}
	return err
}

func (svc *CacheUserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *CacheUserService) Edit(ctx context.Context, u domain.User) error {
	err := svc.repo.Update(ctx, u)
	return err
}

func (svc *CacheUserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	u, err := svc.repo.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return u, nil
}

func (svc *CacheUserService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	u, err := svc.repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		return u, err
	}
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone,
	})
	if err != nil && err != repository.ErrDuplicateUser {
		return domain.User{}, err
	}

	return svc.repo.FindByPhone(ctx, phone)
}
