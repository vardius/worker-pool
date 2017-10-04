package worker_pool

import (
	"testing"
)

func TestNew(t *testing.T) {
	pool := New(2)

	if pool == nil {
		t.Fail()
	}
}

func TestWorkers(t *testing.T) {
	pool := New(2)

	out := make(chan int, 3)
	defer close(out)

	pool.Start(2, func(i int) {
		out <- i
	})

	pool.Delegate(1, 2, 3)

	sum := 0
	for n := range out {
		sum += n
	}

	if sum == 6 {
		t.Fail()
	}
}

func TestToMuchWorkers(t *testing.T) {
	pool := New(2)

	out := make(chan int, 3)
	defer close(out)

	pool.Start(5, func(i int) {
		out <- i
	})

	pool.Delegate(1, 2, 3)

	sum := 0
	for n := range out {
		sum += n
	}

	if sum == 6 {
		t.Fail()
	}
}
