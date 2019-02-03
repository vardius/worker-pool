package workerpool

import (
	"runtime"
	"testing"
)

func poolDelegate(b *testing.B, pool Pool) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		pool.Delegate(n)
	}
}

func poolDelegateParallel(b *testing.B, pool Pool) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.Delegate(1)
		}
	})
}

func runBenchmark(b *testing.B, workersAmount int, runInParallel bool) {
	ch := make(chan int, b.N)
	defer close(ch)

	pool := New(b.N)
	defer pool.Stop()

	pool.Start(workersAmount, func(i int) { ch <- i })

	go func() {
		if runInParallel {
			poolDelegateParallel(b, pool)
		} else {
			poolDelegate(b, pool)
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
