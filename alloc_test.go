package workerpool_test

import (
	"context"
	"testing"

	"github.com/vardius/worker-pool/v3"
)

func noop(int) {}

// TestAllocations guards the allocations per call. Lower a number when an optimisation lands;
// never raise one without a reason in the commit.
func TestAllocations(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector changes allocation counts")
	}

	ctx := context.Background()
	done, cancel := context.WithCancel(ctx)
	cancel()

	const runs = 1000
	room := workerpool.New[int](runs + 1) // AllocsPerRun makes one extra call to warm up
	roomToo := workerpool.New[int](runs + 1)
	stopped := workerpool.New[int](1)
	stopped.Stop()
	finished := make(chan struct{})
	busy := workerpool.New[int](0)
	busy.AddWorker(func(int) { finished <- struct{}{} })
	defer busy.Stop()

	tests := []struct {
		name string
		call func()
		want float64
	}{
		{"Delegate, queue has room", func() { _ = room.Delegate(ctx, 1) }, 0},
		{"Delegate, ctx done", func() { _ = room.Delegate(done, 1) }, 0},
		// The context and its cancel func, WithCancel's. With room, Delegate never asks for its Done channel.
		{"Delegate, queue has room, a new cancellable ctx", func() {
			ctx, cancel := context.WithCancel(ctx)
			_ = roomToo.Delegate(ctx, 1)
			cancel()
		}, 2},
		{"Delegate, pool stopped", func() { _ = stopped.Delegate(ctx, 1) }, 0},
		{"a job through a worker", func() { _ = busy.Delegate(ctx, 1); <-finished }, 0},
		// New: the pool, its queue and its stopping channel; a queue of a T with pointers takes one more,
		// for its buffer. AddWorker: the worker, its quit channel, the closure its go statement makes,
		// and the remove func it returns.
		{"New, AddWorker and Stop", func() {
			p := workerpool.New[int](1)
			p.AddWorker(noop)
			p.Stop()
		}, 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testing.AllocsPerRun(runs, tt.call); got != tt.want {
				t.Errorf("allocates %v times per call, want %v", got, tt.want)
			}
		})
	}
}
