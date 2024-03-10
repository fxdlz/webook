package gormx

import (
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
	"time"
)

type CallBacks struct {
	vector *prometheus.SummaryVec
}

func NewCallBacks(opts prometheus.SummaryOpts) *CallBacks {
	vector := prometheus.NewSummaryVec(opts, []string{"type", "table"})
	prometheus.MustRegister(vector)
	return &CallBacks{
		vector: vector,
	}
}

func (cb *CallBacks) Name() string {
	return "prometheus"
}

func (cb *CallBacks) Initialize(db *gorm.DB) error {
	err := db.Callback().Query().Before("*").Register("prometheus_query_before", cb.Before())
	if err != nil {
		return err
	}
	err = db.Callback().Query().After("*").Register("prometheus_query_after", cb.After("QUERY"))
	if err != nil {
		return err
	}
	err = db.Callback().Create().Before("*").Register("prometheus_create_before", cb.Before())
	if err != nil {
		return err
	}
	err = db.Callback().Create().After("*").Register("prometheus_create_after", cb.After("CREATE"))
	if err != nil {
		return err
	}
	err = db.Callback().Delete().Before("*").Register("prometheus_delete_before", cb.Before())
	if err != nil {
		return err
	}
	err = db.Callback().Delete().After("*").Register("prometheus_delete_after", cb.After("DELETE"))
	if err != nil {
		return err
	}
	err = db.Callback().Update().Before("*").Register("prometheus_update_before", cb.Before())
	if err != nil {
		return err
	}
	err = db.Callback().Update().After("*").Register("prometheus_update_after", cb.After("UPDATE"))
	if err != nil {
		return err
	}
	err = db.Callback().Row().Before("*").Register("prometheus_row_before", cb.Before())
	if err != nil {
		return err
	}
	err = db.Callback().Row().After("*").Register("prometheus_row_after", cb.After("ROW"))
	if err != nil {
		return err
	}
	err = db.Callback().Raw().Before("*").Register("prometheus_raw_before", cb.Before())
	if err != nil {
		return err
	}
	err = db.Callback().Raw().After("*").Register("prometheus_raw_after", cb.After("RAW"))
	return err
}

func (cb *CallBacks) Before() func(*gorm.DB) {
	return func(db *gorm.DB) {
		start := time.Now()
		db.Set("start_time", start)
	}
}

func (cb *CallBacks) After(tp string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		start, ok := val.(time.Time)
		if ok {
			duration := time.Since(start).Milliseconds()
			cb.vector.WithLabelValues(tp, db.Statement.Table).
				Observe(float64(duration))
		}
	}
}
