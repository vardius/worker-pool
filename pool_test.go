package workerpool_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/vardius/worker-pool/v3"
)

// recorder is a worker func that records the jobs it ran.
type recorder struct {
	mu   sync.Mutex
	jobs []int
}

func (r *recorder) run(job int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs = append(r.jobs, job)
}

// ran returns the jobs run so far, sorted.
func (r *recorder) ran() []int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Sorted(slices.Values(r.jobs))
}

// gate returns a worker func that records each job and then waits until release is closed.
func gate(r *recorder, release <-chan struct{}) func(int) {
	return func(job int) {
		r.run(job)
		<-release
	}
}

func delegate(t *testing.T, p *workerpool.Pool[int], jobs ...int) {
	t.Helper()
	for _, job := range jobs {
		if err := p.Delegate(t.Context(), job); err != nil {
			t.Fatalf("Delegate(%d) error = %v", job, err)
		}
	}
}

func TestEveryJobRunsOnce(t *testing.T) {
	for _, queueLength := range []int{0, 1, 16} {
		for _, workers := range []int{1, 3, 32} {
			t.Run(fmt.Sprintf("queue %d, %d workers", queueLength, workers), func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					p := workerpool.New[int](queueLength)
					var r recorder
					for range workers {
						p.AddWorker(r.run)
					}

					want := make([]int, 20)
					for i := range want {
						want[i] = i
					}
					delegate(t, p, want...)
					p.Stop()

					if got := r.ran(); !slices.Equal(got, want) {
						t.Errorf("ran %v, want %v", got, want)
					}
				})
			})
		}
	}
}

func TestJobsWaitForAWorker(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](3)
		delegate(t, p, 1, 2, 3)

		var r recorder
		p.AddWorker(r.run)
		synctest.Wait()

		if got, want := r.ran(), []int{1, 2, 3}; !slices.Equal(got, want) {
			t.Errorf("ran %v, want %v", got, want)
		}
		p.Stop()
	})
}

func TestWorkersRunTheirOwnFunc(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](0)
		release := make(chan struct{})
		var a, b recorder
		p.AddWorker(gate(&a, release))
		p.AddWorker(gate(&b, release))

		delegate(t, p, 1, 2) // each worker takes one and waits
		synctest.Wait()
		close(release)
		p.Stop()

		got := slices.Sorted(slices.Values(slices.Concat(a.ran(), b.ran())))
		if len(a.ran()) != 1 || len(b.ran()) != 1 || !slices.Equal(got, []int{1, 2}) {
			t.Errorf("worker a ran %v and b ran %v, want one of [1 2] each", a.ran(), b.ran())
		}
	})
}

func TestDelegateWaitsForRoom(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](1)
		release := make(chan struct{})
		var r recorder
		p.AddWorker(gate(&r, release))
		delegate(t, p, 1) // the worker takes it and waits
		synctest.Wait()
		delegate(t, p, 2) // fills the queue

		returned := make(chan error)
		go func() { returned <- p.Delegate(t.Context(), 3) }()
		synctest.Wait()
		select {
		case err := <-returned:
			t.Fatalf("Delegate() = %v with the queue full, want it to wait", err)
		default:
		}

		close(release)
		if err := <-returned; err != nil {
			t.Fatalf("Delegate() error = %v once there was room", err)
		}
		p.Stop()
		if got, want := r.ran(), []int{1, 2, 3}; !slices.Equal(got, want) {
			t.Errorf("ran %v, want %v", got, want)
		}
	})
}

func TestDelegateGivesUpWhenCtxIsDone(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](1)
		delegate(t, p, 1) // no workers: the queue is full

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		start := time.Now()
		if err := p.Delegate(ctx, 2); !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Delegate() error = %v, want %v", err, context.DeadlineExceeded)
		}
		if elapsed := time.Since(start); elapsed != time.Second {
			t.Errorf("Delegate() returned after %v, want 1s", elapsed)
		}

		var r recorder
		p.AddWorker(r.run)
		p.Stop()
		if got, want := r.ran(), []int{1}; !slices.Equal(got, want) {
			t.Errorf("ran %v, want %v: the job Delegate gave up on must not run", got, want)
		}
	})
}

func TestDelegateWithCtxDoneQueuesNothing(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](1)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		if err := p.Delegate(ctx, 1); !errors.Is(err, context.Canceled) {
			t.Errorf("Delegate() error = %v, want %v", err, context.Canceled)
		}

		var r recorder
		p.AddWorker(r.run)
		p.Stop()
		if got := r.ran(); len(got) != 0 {
			t.Errorf("ran %v, want nothing: the queue had room, but ctx was done", got)
		}
	})
}

func TestUnbufferedQueueHandsJobsToAWorker(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](0)

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		if err := p.Delegate(ctx, 1); !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Delegate() with no worker: error = %v, want %v", err, context.DeadlineExceeded)
		}

		var r recorder
		p.AddWorker(r.run)
		delegate(t, p, 2)
		p.Stop()
		if got, want := r.ran(), []int{2}; !slices.Equal(got, want) {
			t.Errorf("ran %v, want %v", got, want)
		}
	})
}

// Regression: v2's Delegate after Stop queued the job, which never ran, and hung once the queue was full.
func TestDelegateAfterStop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](1)
		var r recorder
		p.AddWorker(r.run)
		p.Stop()

		for job := range 3 {
			if err := p.Delegate(t.Context(), job); !errors.Is(err, workerpool.ErrStopped) {
				t.Errorf("Delegate(%d) after Stop: error = %v, want %v", job, err, workerpool.ErrStopped)
			}
		}
		if got := r.ran(); len(got) != 0 {
			t.Errorf("ran %v after Stop, want nothing", got)
		}
	})
}

func TestStopWakesDelegateWaitingForRoom(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](1)
		delegate(t, p, 1) // no workers: the queue is full

		returned := make(chan error)
		go func() { returned <- p.Delegate(t.Context(), 2) }()
		synctest.Wait()
		p.Stop()

		if err := <-returned; !errors.Is(err, workerpool.ErrStopped) {
			t.Errorf("Delegate() error = %v, want %v", err, workerpool.ErrStopped)
		}
	})
}

// Regression: v2 had no way to wait for queued jobs; its Stop left them in the queue.
func TestStopRunsQueuedJobsFirst(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](5)
		release := make(chan struct{})
		var r recorder
		p.AddWorker(gate(&r, release))
		delegate(t, p, 1, 2, 3, 4, 5, 6)

		var stopped atomic.Bool
		go func() {
			p.Stop()
			stopped.Store(true)
		}()
		synctest.Wait()
		if stopped.Load() {
			t.Fatal("Stop returned while jobs were still queued")
		}
		if err := p.Delegate(t.Context(), 7); !errors.Is(err, workerpool.ErrStopped) {
			t.Errorf("Delegate() while stopping: error = %v, want %v", err, workerpool.ErrStopped)
		}

		close(release)
		synctest.Wait()
		if !stopped.Load() {
			t.Fatal("Stop didn't return once the queue was done")
		}
		if got, want := r.ran(), []int{1, 2, 3, 4, 5, 6}; !slices.Equal(got, want) {
			t.Errorf("ran %v, want %v", got, want)
		}
		if n := p.WorkersNum(); n != 0 {
			t.Errorf("WorkersNum() = %d after Stop, want 0", n)
		}
	})
}

func TestStopWaitsForRemovedWorkerToFinish(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](0)
		release := make(chan struct{})
		var r recorder
		remove := p.AddWorker(gate(&r, release))
		delegate(t, p, 1)
		synctest.Wait()
		remove()

		var stopped atomic.Bool
		go func() {
			p.Stop()
			stopped.Store(true)
		}()
		synctest.Wait()
		if stopped.Load() {
			t.Fatal("Stop returned while a removed worker was still running a job")
		}
		close(release)
		synctest.Wait()
		if !stopped.Load() {
			t.Fatal("Stop didn't return once the job was done")
		}
	})
}

func TestStopWithNoWorkerDropsQueuedJobs(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](2)
		var r recorder
		p.AddWorker(r.run)() // added and removed at once
		delegate(t, p, 1, 2)

		p.Stop() // returns: nothing is left to run the jobs

		remove := p.AddWorker(r.run)
		synctest.Wait()
		remove()
		if got := r.ran(); len(got) != 0 {
			t.Errorf("ran %v, want nothing", got)
		}
		if n := p.WorkersNum(); n != 0 {
			t.Errorf("WorkersNum() = %d after AddWorker on a stopped pool, want 0", n)
		}
	})
}

// Regression: v2's second Stop closed the quit channels again and panicked.
func TestStopTwice(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](1)
		release := make(chan struct{})
		var r recorder
		p.AddWorker(gate(&r, release))
		delegate(t, p, 1)

		var stopped atomic.Int32
		for range 2 {
			go func() {
				p.Stop()
				stopped.Add(1)
			}()
		}
		synctest.Wait()
		if n := stopped.Load(); n != 0 {
			t.Fatalf("%d Stop calls returned while a job was running, want 0", n)
		}
		close(release)
		synctest.Wait()
		if n := stopped.Load(); n != 2 {
			t.Fatalf("%d Stop calls returned, want 2", n)
		}
		p.Stop()
	})
}

func TestWorkersNum(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](0)
		var r recorder
		removeA := p.AddWorker(r.run)
		removeB := p.AddWorker(r.run)
		p.AddWorker(r.run)

		steps := []struct {
			name string
			do   func()
			want int
		}{
			{"three added", func() {}, 3},
			{"one removed", removeA, 2},
			{"the same one removed again", removeA, 2},
			{"another removed", removeB, 1},
			{"stopped", p.Stop, 0},
			{"removed after Stop", removeB, 0},
		}
		for _, step := range steps {
			step.do()
			if got := p.WorkersNum(); got != step.want {
				t.Errorf("%s: WorkersNum() = %d, want %d", step.name, got, step.want)
			}
		}
	})
}

type handler struct{ r *recorder }

func (h handler) handle(job int) { h.r.run(job) }

// Regression: v2 found workers by their func, so it couldn't remove one added as a method value,
// and returned nil for a func it didn't know.
func TestRemoveTakesOutItsOwnWorker(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](0)
		h := handler{r: &recorder{}}
		removeFirst := p.AddWorker(h.handle)
		p.AddWorker(h.handle)

		removeFirst()
		if n := p.WorkersNum(); n != 1 {
			t.Fatalf("WorkersNum() = %d, want 1", n)
		}
		delegate(t, p, 1, 2)
		p.Stop()
		if got, want := h.r.ran(), []int{1, 2}; !slices.Equal(got, want) {
			t.Errorf("ran %v, want %v", got, want)
		}
	})
}

func TestRemovedWorkerTakesNoNewJob(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](1)
		release := make(chan struct{})
		var removed recorder
		remove := p.AddWorker(gate(&removed, release))
		delegate(t, p, 1) // the worker takes it and waits
		synctest.Wait()
		delegate(t, p, 2) // waits in the queue

		remove()
		close(release)
		synctest.Wait()

		var other recorder
		p.AddWorker(other.run)
		p.Stop()
		if got, want := removed.ran(), []int{1}; !slices.Equal(got, want) {
			t.Errorf("removed worker ran %v, want %v", got, want)
		}
		if got, want := other.ran(), []int{2}; !slices.Equal(got, want) {
			t.Errorf("other worker ran %v, want %v", got, want)
		}
	})
}

func TestRemovedIdleWorkerExits(t *testing.T) {
	// The bubble fails with a deadlock if the removed worker's goroutine is still waiting for a job.
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](1)
		var r recorder
		remove := p.AddWorker(r.run)
		synctest.Wait()
		remove()
	})
}

// Regression: v2 panicked closing a quit channel that Stop had already closed.
func TestRemoveAfterStop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := workerpool.New[int](1)
		var r recorder
		remove := p.AddWorker(r.run)
		p.Stop()
		remove()
	})
}

func TestPanics(t *testing.T) {
	tests := []struct {
		name string
		call func()
		want string
	}{
		{"negative queue length", func() { workerpool.New[int](-1) }, "workerpool: negative queue length"},
		{"nil worker func", func() { workerpool.New[int](1).AddWorker(nil) }, "workerpool: nil worker func"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != tt.want {
					t.Errorf("panic = %v, want %q", got, tt.want)
				}
			}()
			tt.call()
		})
	}
}

// Regression: v2's Delegate and WorkersNum read the workers map while AddWorker wrote it.
// Run with -race. Every job that Delegate accepted must run, even when Stop races with it.
func TestConcurrentUse(t *testing.T) {
	p := workerpool.New[int](4)
	var ran atomic.Int64
	work := func(int) { ran.Add(1) }
	p.AddWorker(work) // never removed, so accepted jobs always have a worker

	var accepted atomic.Int64
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for job := 0; ; job++ {
				err := p.Delegate(t.Context(), job)
				if errors.Is(err, workerpool.ErrStopped) {
					return
				}
				if err != nil {
					t.Errorf("Delegate() error = %v", err)
					return
				}
				accepted.Add(1)
			}
		})
	}
	wg.Go(func() {
		for range 100 {
			remove := p.AddWorker(work)
			_ = p.WorkersNum()
			remove()
		}
	})
	for accepted.Load() < 1000 {
		_ = p.WorkersNum()
	}
	p.Stop()
	wg.Wait()

	if got, want := ran.Load(), accepted.Load(); got != want {
		t.Errorf("%d jobs ran, want the %d that Delegate accepted", got, want)
	}
}
