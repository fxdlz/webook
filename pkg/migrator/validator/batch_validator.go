package validator

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
	"webook/pkg/logger"
	"webook/pkg/migrator"
	"webook/pkg/migrator/events"
)

type BatchValidator[T migrator.Entity] struct {
	base      *gorm.DB
	target    *gorm.DB
	producer  events.Producer
	l         logger.LoggerV1
	direction string
	batchSize int
}

func NewBatchValidator[T migrator.Entity](
	base *gorm.DB,
	target *gorm.DB,
	direction string,
	l logger.LoggerV1,
	p events.Producer) *BatchValidator[T] {
	res := &BatchValidator[T]{base: base, target: target,
		l: l, producer: p, direction: direction, batchSize: 100}
	return res
}

func (v *BatchValidator[T]) batchFullFromBase(ctx context.Context, offset, limit int) ([]T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src []T
	err := v.base.WithContext(dbCtx).Order("id").Offset(offset).Limit(limit).Find(&src).Error
	return src, err
}

// 批量校验
func (v *BatchValidator[T]) batchValidateBaseToTarget(ctx context.Context) error {
	offset := -v.batchSize
	for {
		offset += v.batchSize
		src, err := v.batchFullFromBase(ctx, offset, v.batchSize)
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		if err != nil {
			//记录错误日志
			continue
		}
		var srcMapIds map[int64]T
		for _, item := range src {
			srcMapIds[item.ID()] = item
		}
		ids := slice.Map[T, int64](src, func(idx int, t T) int64 {
			return t.ID()
		})
		var dst []T
		err = v.target.WithContext(ctx).Where("id IN ?", ids).Find(&dst).Error

		switch err {
		case gorm.ErrRecordNotFound:
			for _, id := range ids {
				v.notify(id, events.InconsistentEventTypeTargetMissing)
			}
		case nil:
			for _, t := range dst {
				if !t.CompareTo(srcMapIds[t.ID()]) {
					v.notify(t.ID(), events.InconsistentEventTypeNEQ)
				}
			}
			diff := slice.DiffSetFunc[T](src, dst, func(src, dst T) bool {
				return src.ID() == dst.ID()
			})
			v.notifyBaseMissing(diff)
		default:
			// 记录日志，然后继续
			// 做好监控
			v.l.Error("base -> target 查询 target 失败",
				logger.Error(err))
		}
	}
}

func (v *BatchValidator[T]) notifyBaseMissing(ts []T) {
	for _, val := range ts {
		v.notify(val.ID(), events.InconsistentEventTypeBaseMissing)
	}
}

func (v *BatchValidator[T]) notify(id int64, typ string) {
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
