package failover

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"sync/atomic"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/internal/service/sms"
	"webook/internal/service/sms/ratelimit"
)

type CodeFailOverSMSService struct {
	svcs        []sms.Service
	repo        repository.ReqRetryRepository
	idx         int32
	retryNum    int8
	interval    int32
	reqErrCount int32
	reqAllCount int32
}

type SMSReq struct {
	tplId   string
	args    []string
	numbers []string
}

func (c *CodeFailOverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&c.idx)
	err := c.svcs[idx].Send(ctx, tplId, args, numbers...)
	if err != nil {
		if err == ratelimit.ErrLimited {
			id, saveErr := c.saveReqInDB(ctx, tplId, args, numbers)
			if saveErr != nil {
				return saveErr
			}
			go c.retry(ctx, id)
		} else {
			reqErr := atomic.LoadInt32(&c.reqErrCount)
			reqAll := atomic.LoadInt32(&c.reqAllCount)
			if reqErr > 0 && reqAll > 10 && reqAll/reqErr >= 2 {
				id, saveErr := c.saveReqInDB(ctx, tplId, args, numbers)
				if saveErr != nil {
					return saveErr
				}
				atomic.StoreInt32(&c.reqErrCount, 0)
				newIdx := (idx + 1) % int32(len(c.svcs))
				atomic.CompareAndSwapInt32(&c.idx, idx, newIdx)
				go c.retry(ctx, id)
			}
			atomic.AddInt32(&c.reqErrCount, 1)
		}
		atomic.AddInt32(&c.reqAllCount, 1)
	} else {
		atomic.StoreInt32(&c.reqErrCount, 0)
		atomic.StoreInt32(&c.reqAllCount, 0)
	}
	return err
}

func (c *CodeFailOverSMSService) saveReqInDB(ctx context.Context, tplId string, args []string, numbers []string) (string, error) {
	newUUID, uuidErr := uuid.NewUUID()
	if uuidErr != nil {
		return "", uuidErr
	}
	id := newUUID.String()
	data, jsonErr := json.Marshal(SMSReq{
		tplId:   tplId,
		args:    args,
		numbers: numbers,
	})
	if jsonErr != nil {
		return "", jsonErr
	}
	createErr := c.repo.Create(ctx, domain.ReqRetry{Id: id, Req: string(data)})
	if createErr != nil {
		return "", createErr
	}
	return id, nil
}

func (c *CodeFailOverSMSService) retry(ctx context.Context, id string) {
	r, err := c.repo.FindById(ctx, id)
	var req SMSReq
	err = json.Unmarshal([]byte(r.Req), &req)
	if err != nil {
		//打印日志
		return
	}
	var count int8 = 0
	for count < c.retryNum {
		time.Sleep(time.Second * time.Duration(c.interval))
		count++
		idx := atomic.LoadInt32(&c.idx)
		err = c.svcs[idx].Send(ctx, req.tplId, req.args, req.numbers...)
		if err == nil {
			break
		}
	}
	err = c.repo.Delete(ctx, id)
	if err != nil {
		//打印日志
	}
}

func NewCodeFailOverSMSService(svcs []sms.Service, repo repository.ReqRetryRepository) *CodeFailOverSMSService {
	return &CodeFailOverSMSService{
		svcs: svcs,
		repo: repo,
	}
}
