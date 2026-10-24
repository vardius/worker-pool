// Package workerpool runs jobs on a pool of goroutines fed by one bounded queue.
//
// A Pool[T] queues values of type T. Each worker is a goroutine that takes jobs from the queue
// one at a time and calls its func with them:
//
//	pool := workerpool.New[string](100)
//	for range runtime.NumCPU() {
//		pool.AddWorker(resize) // resize is a func(path string)
//	}
//
//	for _, path := range paths {
//		if err := pool.Delegate(ctx, path); err != nil {
//			return err
//		}
//	}
//	pool.Stop() // runs what is still queued, then returns
//
// # Queue
//
// New takes the queue's length. Delegate queues a job and returns at once while there is room.
// When the queue is full, it waits for a worker to take a job, and gives up with ctx's error
// when ctx is done first: pass a context with a deadline to bound the wait. Jobs wait in the queue
// until a worker is free, so they can be queued before any worker is added.
//
// # Workers
//
// AddWorker starts a worker and returns a func that removes it. A removed worker finishes the job
// it is running and exits. Workers can be added and removed at any time, and each can run its own func.
// A panic in a worker's func is not recovered, as in any goroutine.
//
// # Stopping
//
// Stop stops the pool taking jobs: from then on Delegate returns ErrStopped. The workers run every
// job still in the queue, and Stop returns once they have all exited, so nothing runs after it returns.
// Calling Stop again does nothing but wait for the same. To shut down faster, have the jobs watch a
// context and cancel it before calling Stop.
//
// # Cost
//
// Delegate into a queue with room and the worker loop allocate nothing: a job is a typed value sent on a channel,
// with no reflection and no boxing. Delegate waiting for room asks ctx for its Done channel, which a cancellable
// context makes, once, the first time it is asked. AddWorker allocates the worker, its quit channel, its goroutine's closure
// and the remove func it returns.
package workerpool
