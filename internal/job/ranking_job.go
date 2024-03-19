package job

import (
	"context"
	_ "embed"
	rlock "github.com/gotomicro/redis-lock"
	uuid "github.com/lithammer/shortuuid/v4"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"sync"
	"time"
	"webook/internal/service"
	"webook/pkg/logger"
)

//go:embed lua/update_load.lua
var luaUpdateLoad string

type RankingJob struct {
	svc       service.RankingService
	l         logger.LoggerV1
	timeout   time.Duration
	client    *rlock.Client
	key       string
	localLock *sync.Mutex
	lock      *rlock.Lock
	cmd       redis.Cmdable
	load      int
	node      string
}

func NewRankingJob(svc service.RankingService, timeout time.Duration, l logger.LoggerV1, client *rlock.Client) *RankingJob {
	return &RankingJob{
		svc:       svc,
		timeout:   timeout,
		l:         l,
		client:    client,
		localLock: &sync.Mutex{},
		key:       "job:ranking",
		node:      uuid.New(),
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Key() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	go func() {
		//模拟节点负载
		for {
			rd := rand.New(rand.NewSource(time.Now().UnixMilli()))
			r.load = rd.Intn(100)
			time.Sleep(time.Second * time.Duration(r.load))
		}
	}()
	r.localLock.Lock()
	lock := r.lock
	if lock == nil {
		//尝试更新最小负载节点
		res, err1 := r.cmd.Eval(context.Background(), luaUpdateLoad, []string{r.key}, r.node, r.load, r.timeout).Int()
		if err1 != nil {
			r.localLock.Unlock()
			r.l.Warn("lua脚本执行失败")
			return err1
		}
		if res == 1 {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
				Interval: time.Millisecond * 100,
				Max:      3,
			}, time.Second)
			if err != nil {
				r.localLock.Unlock()
				r.l.Warn("获取分布式锁失败", logger.Error(err))
				return nil
			}
			r.lock = lock
			r.localLock.Unlock()
			go func() {
				er := lock.AutoRefresh(r.timeout/2, r.timeout)
				if er != nil {
					r.localLock.Lock()
					r.lock = nil
					r.localLock.Unlock()
				}
			}()
		} else {
			r.localLock.Unlock()
			r.l.Info("该节点负载不是最低，由其他节点执行任务", logger.String("node", r.node))
			return nil
		}
	} else {
		//尝试更新最小负载节点
		res, err1 := r.cmd.Eval(context.Background(), luaUpdateLoad, []string{r.key}, r.node, r.load, r.timeout).Int()
		if err1 != nil {
			r.l.Warn("lua脚本执行失败")
		} else {
			//更新失败证明该节点当前不是负载最小节点，释放分布式锁
			if res == -1 {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := r.lock.Unlock(ctx)
				if err != nil {
					r.l.Warn("解除分布式锁失败", logger.Error(err))
				} else {
					r.localLock.Lock()
					r.lock = nil
					r.localLock.Unlock()
				}
			}
		}
	}
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}
