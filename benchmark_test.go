package workerpool

import "testing"

func BenchmarkWorker(b *testing.B) {
	pool := New(2)
	defer pool.Stop()

	pool.Start(2, func(i int) {})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.Delegate(1)
		}
	})
}
