package workerpool

import (
	"runtime"
	"testing"
)

func poolDelegate(b *testing.B, pool Pool, out chan<- int) {
	for n := 0; n < b.N; n++ {
		pool.Delegate(n, out)
	}
}

func poolDelegateParallel(b *testing.B, pool Pool, out chan<- int) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.Delegate(1, out)
		}
	})
}

func runBenchmark(b *testing.B, workersAmount int, runInParallel bool) {
	ch := make(chan int, b.N)
	defer close(ch)

	pool := New(b.N)
	defer pool.Stop()

	worker := func(i int, out chan<- int) { out <- i }

	for i := 1; i <= workersAmount; i++ {
		if err := pool.AddWorker(worker); err != nil {
			b.Fatal(err)
		}
	}

	go func() {
		if runInParallel {
			poolDelegateParallel(b, pool, ch)
		} else {
			poolDelegate(b, pool, ch)
		}
	}()

	var i = 0
	for i < b.N {
		select {
		case <-ch:
			i++
		}
	}
}

func BenchmarkWorker1(b *testing.B) {
	runBenchmark(b, 1, false)
}

func BenchmarkWorker1Parallel(b *testing.B) {
	runBenchmark(b, 1, true)
}

func BenchmarkWorker100(b *testing.B) {
	runBenchmark(b, 100, false)
}

func BenchmarkWorker100Parallel(b *testing.B) {
	runBenchmark(b, 100, true)
}

func BenchmarkWorkerNumCPU(b *testing.B) {
	runBenchmark(b, runtime.NumCPU()+1, false)
}

func BenchmarkWorkerNumCPUParallel(b *testing.B) {
	runBenchmark(b, runtime.NumCPU()+1, true)
}
