package cronjob

import (
	"github.com/ecodeclub/ekit/queue"
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

func TestPQ(t *testing.T) {
	pq := queue.NewPriorityQueue[int](5, func(a int, b int) int {
		if a > b {
			return -1
		} else if a < b {
			return 1
		} else {
			return 0
		}
	})
	pq.Enqueue(5)
	pq.Enqueue(3)
	pq.Enqueue(4)
	pq.Enqueue(1)
	pq.Enqueue(2)
	pq.Enqueue(6)
	pq.Enqueue(7)
	for pq.Len() > 0 {
		t.Log(pq.Peek())
		pq.Dequeue()
	}
}
