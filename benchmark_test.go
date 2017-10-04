package workerpool

import (
	"bytes"
	"testing"
)

func BenchmarkWorker(b *testing.B) {
	pool := New(2)

	out := make(chan int, 3)
	defer close(out)

	pool.Start(2, func(i int) {
		out <- i
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var buf bytes.Buffer
		for pb.Next() {
			buf.Reset()
			pool.Delegate(1)
		}
	})
}
