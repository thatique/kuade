package queue

import (
	"fmt"

	"testing"
	"time"
)

type testJob struct {
	s string
}

func (t *testJob) GetName() string {
	return fmt.Sprintf("testjob-%v", t.s)
}

func (t *testJob) Fire() error {
	return nil
}

func TestBasicOperation(t *testing.T) {
	queue := NewQueue(10, 300)
	var executed int
	queue.SetOnExpiredJob(func(jobName string, now time.Time) bool {
		executed++
		return true
	})
	for i := 0; i < 100; i++ {
		for j := 0; j < 20; j++ {
			go queue.Push(&testJob{s: fmt.Sprintf("push-%v-%v", i, j)})
		}

		go queue.Later(time.Second/2, &testJob{s: fmt.Sprintf("later-%v", i)})
	}

	time.Sleep(time.Millisecond * 1500)
	n := queue.Len()
	if n != 0 {
		t.Fatalf("expecting queue len equal to '%v', got '%v'", 0, n)
	}

	if executed != 100 {
		t.Fatalf("expecting delayed job executed equal to '%v', got '%v'", 0, n)
	}

	queue.Stop()
}
