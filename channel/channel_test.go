package channel

import (
	"testing"
)

func TestChannel(t *testing.T) {

	ch1 := make(chan int)
	ch2 := make(chan int)
	go func() {
		ch1 <- 1
		close(ch1)
	}()
	go func() {
		ch2 <- 1
		close(ch2)
	}()
	select {
	case val := <-ch1:
		t.Log("进来了ch1", val)
	case val := <-ch2:
		t.Log("进来了ch2", val)
	}
}
