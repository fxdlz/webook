package job

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"time"
	"webook/internal/domain"
	"webook/internal/service"
	"webook/pkg/logger"
)

type Executor interface {
	Name() string
	Exec(ctx context.Context, job domain.Job) error
}

type LocalExecutor struct {
	funcs map[string]func(ctx context.Context, job domain.Job) error
}

func (l *LocalExecutor) RegisterFunc(name string, fn func(ctx context.Context, job domain.Job) error) {
	l.funcs[name] = fn
}

func (l *LocalExecutor) Name() string {
	return "local_executor"
}

func (l *LocalExecutor) Exec(ctx context.Context, job domain.Job) error {
	jobFunc, ok := l.funcs[job.Name]
	if !ok {
		return fmt.Errorf("未注册本地方法 %s", job.Name)
	}
	return jobFunc(ctx, job)
}

type Scheduler struct {
	dbTimeout time.Duration
	svc       service.CronJobService
	executors map[string]Executor
	l         logger.LoggerV1
	limiter   semaphore.Weighted
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.executors[exec.Name()] = exec
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		dbCtx, cancel := context.WithTimeout(ctx, s.dbTimeout)
		job, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			continue
		}
		executor, ok := s.executors[job.Executor]
		if !ok {
			s.l.Warn("找不到执行器", logger.Int64("jid", job.Id), logger.String("executor", job.Executor))
			continue
		}
		go func() {
			defer func() {
				s.limiter.Release(1)
				job.CancelFunc()
			}()
			er := executor.Exec(ctx, job)
			if er != nil {
				s.l.Warn("执行任务失败", logger.Error(er), logger.Int64("jid", job.Id))
				return
			}
			er = s.svc.ResetNextTime(ctx, job)
			if er != nil {
				s.l.Warn("刷新下一次执行时间失败", logger.Error(er), logger.Int64("jid", job.Id))
			}
		}()
	}
}
