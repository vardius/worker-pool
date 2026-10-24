package workerpool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// ErrStopped is what Delegate returns once the pool is stopped.
var ErrStopped = errors.New("workerpool: pool stopped")

// Pool queues jobs of type T and runs them on the workers added with AddWorker.
// Create one with New. Its methods are safe to call from several goroutines.
type Pool[T any] struct {
	queue    chan T
	stopping chan struct{} // closed when Stop starts, to wake Delegate calls waiting for room

	// Delegate and AddWorker hold mu for reading, so Stop, which takes it for writing,
	// closes the queue only once nothing can send to it or start a worker any more.
	mu      sync.RWMutex
	stopped bool

	workers  atomic.Int64   // added and not yet removed
	running  sync.WaitGroup // every worker goroutine, removed ones that are finishing a job included
	stopOnce sync.Once
}

// New returns a pool whose queue holds up to queueLength jobs. With a queueLength of 0,
// Delegate waits until a worker takes the job. The pool has no workers until AddWorker adds them.
// New panics if queueLength is negative.
func New[T any](queueLength int) *Pool[T] {
	if queueLength < 0 {
		panic("workerpool: negative queue length")
	}
	return &Pool[T]{queue: make(chan T, queueLength), stopping: make(chan struct{})}
}

// Delegate queues job for the next free worker. If the queue is full, it waits for room until
// ctx is done, returning ctx's error, or the pool is stopped, returning ErrStopped.
// If ctx is already done, or the pool is stopped, it returns at once and doesn't queue job.
//
// A job that Delegate queued, with a nil error, runs before Stop returns, as long as a worker is left to run it;
// Stop drops it otherwise.
func (p *Pool[T]) Delegate(ctx context.Context, job T) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.stopped {
		return ErrStopped
	}

	// A plain send while there is room: the select below locks each of its channels,
	// and ctx.Done makes a cancellable context allocate its channel.
	select {
	case p.queue <- job:
		return nil
	default:
	}
	select {
	case p.queue <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-p.stopping:
		return ErrStopped
	}
}

// AddWorker starts a worker: a goroutine that takes jobs from the queue one at a time and calls fn with each.
// A panic in fn is not recovered.
//
// It returns a func that removes the worker. A removed worker finishes the job it is running, if any, and exits;
// one that is between jobs when remove is called may still take one more. remove doesn't wait for the worker.
// Calling remove again does nothing. If fn calls runtime.Goexit, the worker exits as if removed.
// After Stop, AddWorker starts nothing and returns a remove that does nothing. AddWorker panics if fn is nil.
func (p *Pool[T]) AddWorker(fn func(job T)) (remove func()) {
	if fn == nil {
		panic("workerpool: nil worker func")
	}

	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.stopped {
		return func() {}
	}

	w := &worker[T]{pool: p, queue: p.queue, fn: fn, quit: make(chan struct{})}
	p.workers.Add(1)
	p.running.Add(1)
	go w.run()

	return w.remove
}

// WorkersNum returns the number of workers added and not removed. Workers stop counting
// once they exit after Stop, so it is 0 when Stop returns.
func (p *Pool[T]) WorkersNum() int {
	return int(p.workers.Load())
}

// Stop stops the pool taking jobs: Delegate returns ErrStopped from then on, and calls already
// waiting for room return too. The workers run every job still queued, and Stop returns once every
// worker has exited, removed ones finishing their last job included. Jobs still queued when no worker
// is left are dropped.
//
// Stop can be called more than once, and from several goroutines at the same time: every call returns
// once the workers have exited. Calling it from a job never returns, as it would wait for the job's own worker.
func (p *Pool[T]) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopping) // first, so that Delegate calls waiting for room let go of mu
		p.mu.Lock()
		p.stopped = true
		close(p.queue)
		p.mu.Unlock()
	})
	p.running.Wait()
}

type worker[T any] struct {
	pool  *Pool[T]
	queue <-chan T // the pool's, kept here so the worker's loop never reads the memory that Delegate writes
	fn    func(T)
	quit  chan struct{} // closed by remove
	left  atomic.Bool   // set once the worker no longer counts in WorkersNum
}

func (w *worker[T]) run() {
	defer w.pool.running.Done()
	defer w.leave()

	for {
		job, ok := w.next()
		if !ok {
			return
		}
		w.fn(job)
	}
}

// next waits for a job. It returns false once the worker is removed, or once the pool is stopped
// and its queue is empty.
func (w *worker[T]) next() (job T, ok bool) {
	// A worker removed before this check takes no new job, even with jobs waiting. On an open channel, it takes no lock.
	select {
	case <-w.quit:
		return job, false
	default:
	}
	// A plain receive while jobs are waiting: a select over the queue and quit would lock both.
	select {
	case job, ok = <-w.queue:
		return job, ok
	default:
	}
	select {
	case job, ok = <-w.queue:
		return job, ok
	case <-w.quit:
		return job, false
	}
}

func (w *worker[T]) remove() {
	if w.leave() {
		close(w.quit)
	}
}

// leave takes the worker off the count, the first time it is called.
func (w *worker[T]) leave() bool {
	if w.left.Swap(true) {
		return false
	}
	w.pool.workers.Add(-1)
	return true
}
