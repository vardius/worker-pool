package workerpool

import (
	"fmt"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	pool := New(2)
	defer pool.Stop()

	if pool == nil {
		t.Fail()
	}
}

func TestWorkers(t *testing.T) {
	delegateWorkToWorkers(t, 2, 3, 3) // same workers as jobs
	delegateWorkToWorkers(t, 2, 3, 2) // less workers then jobs
	delegateWorkToWorkers(t, 2, 3, 5) // more workers than jobs
}

func delegateWorkToWorkers(t *testing.T, poolSize int, jobsAmount int, workersAmount int) {
	var wg sync.WaitGroup

	pool := New(poolSize)
	out := make(chan int, jobsAmount)

	go func() {
		defer pool.Stop()
		wg.Wait()
		close(out)
	}()

	pool.Start(2, func(i int) {
		defer wg.Done()
		out <- i
	})

	wg.Add(workersAmount)

	for i := 0; i < jobsAmount; i++ {
		pool.Delegate(i)
	}

	sum := 0
	for n := range out {
		sum += n
	}

	if sum == 0 {
		fmt.Errorf("Delegating job %d to %d workers faild", jobsAmount, workersAmount)
		t.Fail()
	}
}
