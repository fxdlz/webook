package cronjob

import (
	"github.com/robfig/cron/v3"
	"testing"
	"time"
)

func TestCron(t *testing.T) {
	expr := cron.New(cron.WithSeconds())
	_, err := expr.AddFunc("@every 1s", func() {
		time.Sleep(time.Second * 2)
		t.Log("执行中")
	})
	if err != nil {
		panic(err)
	}
	expr.Start()
	time.Sleep(time.Second * 3)
	t.Log("停止继续执行新的定时任务")
	ctx := expr.Stop()
	<-ctx.Done()
	t.Log("没有任务在执行")
}
