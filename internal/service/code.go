package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"
	"webook/internal/repository"
	"webook/internal/service/sms"
)

var ErrCodeSendTooMany = repository.ErrCodeSendTooMany

type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
	generateCode() string
}

type CacheCodeService struct {
	repo repository.CodeRepository
	sms  sms.Service
}

func NewCacheCodeService(repo repository.CodeRepository, sms sms.Service) CodeService {
	return &CacheCodeService{
		repo: repo,
		sms:  sms,
	}
}

func (svc *CacheCodeService) Send(ctx context.Context, biz, phone string) error {
	code := svc.generateCode()
	err := svc.repo.Set(ctx, biz, phone, code)
	//发送验证码
	if err != nil {
		return err
	}
	const codeTplId = "21131212412"
	svc.sms.Send(ctx, codeTplId, []string{code}, phone)
	return err
}

func (svc *CacheCodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	ok, err := svc.repo.Verify(ctx, biz, phone, inputCode)
	if err == repository.ErrCodeVerifyTooMany {
		return false, nil
	}
	return ok, err
}

func (svc *CacheCodeService) generateCode() string {
	rand.Seed(time.Now().UnixMilli())
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}
