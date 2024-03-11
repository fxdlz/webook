package saramax

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"webook/pkg/logger"
)

type Handler[T any] struct {
	l      logger.LoggerV1
	fn     func(msg *sarama.ConsumerMessage, event T) error
	vector *prometheus.SummaryVec
}

func NewHandler[T any](l logger.LoggerV1, opt prometheus.SummaryOpts, fn func(msg *sarama.ConsumerMessage, event T) error) *Handler[T] {
	vec := prometheus.NewSummaryVec(opt, []string{"topic"})
	prometheus.MustRegister(vec)
	return &Handler[T]{
		l:      l,
		vector: vec,
		fn:     fn,
	}
}

func (h *Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()

	for msg := range msgs {
		start := time.Now()
		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			h.l.Error("反序列消息体失败",
				logger.String("topic", msg.Topic),
				logger.Int32("partition", msg.Partition),
				logger.Int64("offset", msg.Offset),
				logger.Error(err))
		} else {
			er := h.fn(msg, t)
			if er != nil {
				h.l.Error("处理消息失败",
					logger.String("topic", msg.Topic),
					logger.Int32("partition", msg.Partition),
					logger.Int64("offset", msg.Offset),
					logger.Error(er))
			}
		}
		session.MarkMessage(msg, "")
		duration := time.Since(start).Milliseconds()
		h.vector.WithLabelValues(msg.Topic).Observe(float64(duration))
	}
	return nil
}
