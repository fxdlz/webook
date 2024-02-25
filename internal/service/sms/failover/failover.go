package failover

import (
	"context"
	"errors"
	"sync/atomic"
	"webook/internal/service/sms"
)

type FailOverSMSService struct {
	svcs []sms.Service
	idx  uint64
}

func NewFailOverSMSService(svcs []sms.Service) *FailOverSMSService {
	return &FailOverSMSService{
		svcs: svcs,
	}
}

func (f *FailOverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	for _, svc := range f.svcs {
		err := svc.Send(ctx, tplId, args, numbers...)
		if err == nil {
			return nil
		}
		//打印日志
	}
	return errors.New("发送失败，所有服务商都尝试过了。")
}

func (f *FailOverSMSService) SendV1(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))
	for i := idx; i < (idx + length); i++ {
		svc := f.svcs[i%length]
		err := svc.Send(ctx, tplId, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.Canceled, context.DeadlineExceeded:
			return err
		}
		//打印日志
	}
	return errors.New("发送失败，所有服务商都尝试过了。")
}
