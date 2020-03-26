package workerpool

import (
	"runtime"
	"testing"
)

func poolDelegate(b *testing.B, pool Pool) {
	for n := 0; n < b.N; n++ {
		if err := pool.Delegate(n); err != nil {
			b.Fatal(err)
		}
	}
}

func poolDelegateParallel(b *testing.B, pool Pool) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := pool.Delegate(1); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func runBenchmark(b *testing.B, workersAmount int, runInParallel bool) {
	pool := New(b.N)
	defer pool.Stop()

	worker := func(i int) {}

	for i := 1; i <= workersAmount; i++ {
		if err := pool.AddWorker(worker); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	if runInParallel {
		poolDelegateParallel(b, pool)
	} else {
		poolDelegate(b, pool)
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
