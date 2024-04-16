package validator

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
	"webook/pkg/logger"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"
)

type Validator[T migrator.Entity] struct {
	base          *gorm.DB
	target        *gorm.DB
	producer      events.Producer
	l             logger.LoggerV1
	direction     string
	batchSize     int
	utime         int64
	sleepInterval time.Duration
	fromBase      func(ctx context.Context, offset int) (T, error)
}

func NewValidator[T migrator.Entity](
	base *gorm.DB,
	target *gorm.DB,
	direction string,
	l logger.LoggerV1,
	p events.Producer) *Validator[T] {
	res := &Validator[T]{base: base, target: target,
		l: l, producer: p, direction: direction, batchSize: 100}
	res.fromBase = res.fullFromBase
	return res
}

func (v *Validator[T]) Utime(t int64) *Validator[T] {
	v.utime = t
	return v
}

func (v *Validator[T]) SleepInterval(interval time.Duration) *Validator[T] {
	v.sleepInterval = interval
	return v
}

func (v *Validator[T]) Full() *Validator[T] {
	v.fromBase = v.fullFromBase
	return v
}

func (v *Validator[T]) Incr() *Validator[T] {
	v.fromBase = v.incrFromBase
	return v
}

func (v *Validator[T]) incrFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).Where("utime > ?", v.utime).
		Order("utime").
		Offset(offset).
		First(&src).Error
	return src, err
}

func (v *Validator[T]) fullFromBase(ctx context.Context, offset int) (T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src T
	err := v.base.WithContext(dbCtx).Order("id").Offset(offset).First(&src).Error
	return src, err
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group
	eg.Go(func() error {
		return v.validateBaseToTarget(ctx)
	})
	eg.Go(func() error {
		return v.validateTargetToBase(ctx)
	})
	return eg.Wait()
}

func (v *Validator[T]) validateBaseToTarget(ctx context.Context) error {
	offset := -1
	for {
		offset++
		src, err := v.fromBase(ctx, offset)
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		if err != nil {
			//记录错误日志
			continue
		}
		var dst T
		err = v.target.WithContext(ctx).Where("id=?", src.ID()).First(&dst).Error
		switch err {
		case gorm.ErrRecordNotFound:
			v.notify(src.ID(), events.InconsistentEventTypeTargetMissing)
		case nil:
			if !src.CompareTo(dst) {
				v.notify(src.ID(), events.InconsistentEventTypeNEQ)
			}
		default:
			// 记录日志，然后继续
			// 做好监控
			v.l.Error("base -> target 查询 target 失败",
				logger.Int64("id", src.ID()),
				logger.Error(err))
		}
	}
}

func (v *Validator[T]) validateTargetToBase(ctx context.Context) error {
	offset := -1
	for {
		offset++
		var ts []T
		err := v.target.WithContext(ctx).Order("id").Offset(offset).Limit(v.batchSize).Find(&ts).Error
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		if err != nil {
			v.l.Error("target => base 查询 target 失败", logger.Error(err))
			offset += len(ts)
			continue
		}
		var srcTs []T
		ids := slice.Map(ts, func(idx int, t T) int64 {
			return t.ID()
		})
		err = v.base.WithContext(ctx).Where("id IN ?", ids).Find(&srcTs).Error
		if err == gorm.ErrRecordNotFound {

		}
		diff := slice.DiffSetFunc[T](ts, srcTs, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		v.notifyBaseMissing(diff)
		offset += len(ts)
	}
}

func (v *Validator[T]) notifyBaseMissing(ts []T) {
	for _, val := range ts {
		v.notify(val.ID(), events.InconsistentEventTypeBaseMissing)
	}
}

func (v *Validator[T]) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := v.producer.ProduceInconsistentEvent(ctx, events.InconsistentEvent{
		ID:        id,
		Type:      typ,
		Direction: v.direction,
	})
	if err != nil {
		v.l.Error("发送不一致消息失败",
			logger.Error(err),
			logger.String("type", typ),
			logger.Int64("id", id))
	}
}
