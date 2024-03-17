package job

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"strconv"
	"time"
	"webook/pkg/logger"
)

type CronJobBuilder struct {
	l      logger.LoggerV1
	vector *prometheus.SummaryVec
}

func NewCronJobBuilder(l logger.LoggerV1, opt prometheus.SummaryOpts) *CronJobBuilder {
	vector := prometheus.NewSummaryVec(opt, []string{"job", "success"})
	prometheus.MustRegister(vector)
	return &CronJobBuilder{l: l, vector: vector}
}

func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobAdapterFunc(func() {
		start := time.Now()
		b.l.Debug("开始执行", logger.String("name", name))
		err := job.Run()
		if err != nil {
			b.l.Error("执行失败", logger.Error(err), logger.String("name", name))
		}
		b.l.Debug("执行结束", logger.String("name", name))
		duration := time.Since(start)
		b.vector.WithLabelValues(name, strconv.FormatBool(err == nil)).Observe(float64(duration.Milliseconds()))
	})
}

type cronJobAdapterFunc func()

func (f cronJobAdapterFunc) Run() {
	f()
}
