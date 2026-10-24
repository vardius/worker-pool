package workerpool_test

import (
	"context"
	"runtime"
	"strconv"
	"testing"

	"github.com/vardius/worker-pool/v3"
)

// The jobs do nothing, so what is measured is the pool's own cost: queueing a job and handing it to a worker.

// BenchmarkDelegate queues jobs with no worker to take them: Delegate alone, into a queue with room.
func BenchmarkDelegate(b *testing.B) {
	const queueLength = 1 << 16
	ctx := context.Background()
	p := workerpool.New[int](queueLength)
	queued := 0
	b.ReportAllocs()
	for b.Loop() {
		if queued == queueLength {
			// Full: carry on with an empty pool, off the clock.
			b.StopTimer()
			p, queued = workerpool.New[int](queueLength), 0
			b.StartTimer()
		}
		_ = p.Delegate(ctx, queued)
		queued++
	}
}

// BenchmarkJob delegates jobs one after another to a pool whose workers do nothing.
// Per op is the time per job, from Delegate to the worker's call.
func BenchmarkJob(b *testing.B) {
	for _, workers := range []int{1, runtime.GOMAXPROCS(0)} {
		b.Run("workers="+strconv.Itoa(workers), func(b *testing.B) {
			p := newBusyPool(workers)
			defer p.Stop()
			ctx := context.Background()
			b.ReportAllocs()
			job := 0
			for b.Loop() {
				_ = p.Delegate(ctx, job)
				job++
			}
		})
	}
}

// BenchmarkJobParallel is BenchmarkJob with GOMAXPROCS goroutines delegating at once.
func BenchmarkJobParallel(b *testing.B) {
	p := newBusyPool(runtime.GOMAXPROCS(0))
	defer p.Stop()
	ctx := context.Background()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		job := 0
		for pb.Next() {
			_ = p.Delegate(ctx, job)
			job++
		}
	})
}

func newBusyPool(workers int) *workerpool.Pool[int] {
	p := workerpool.New[int](1024)
	for range workers {
		p.AddWorker(noop)
	}
	return p
}
