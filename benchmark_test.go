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

func BenchmarkWorkerNumCPU(b *testing.B) {
	pool := New(runtime.NumCPU())
	defer pool.Stop()

	pool.Start(runtime.NumCPU(), func(i int) {})

	poolDelegate(b, pool)
}

func BenchmarkWorkerNumCPUParallel(b *testing.B) {
	pool := New(runtime.NumCPU())
	defer pool.Stop()

	pool.Start(runtime.NumCPU(), func(i int) {})

	poolDelegateParallel(b, pool)
}

func BenchmarkWorker(b *testing.B) {
	pool := New(1)
	defer pool.Stop()

	pool.Start(1, func(i int) {})

	poolDelegate(b, pool)
}

func BenchmarkWorkerParallel(b *testing.B) {
	pool := New(1)
	defer pool.Stop()

	pool.Start(1, func(i int) {})

	poolDelegateParallel(b, pool)
}

func BenchmarkWorker100(b *testing.B) {
	pool := New(100)
	defer pool.Stop()

	pool.Start(100, func(i int) {})

	poolDelegate(b, pool)
}

func BenchmarkWorker100Parallel(b *testing.B) {
	pool := New(100)
	defer pool.Stop()

	pool.Start(100, func(i int) {})

	poolDelegateParallel(b, pool)
}
