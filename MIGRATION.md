# Migrating from v2 to v3

v3 keeps the idea: a bounded queue of jobs and workers that take them. The pool is now generic, so a job is a
typed value and a worker is a `func(T)`, with no reflection. A worker is removed with the func `AddWorker`
returns instead of by its func value, `Delegate` takes a context, and `Stop` runs the queued jobs and waits for
the workers. The edge cases that changed are listed below.

## 1. Update the import path

```sh
go get github.com/vardius/worker-pool/v3
```

```go
import "github.com/vardius/worker-pool/v3"
```

The package is still called `workerpool`. v3 needs Go 1.26 or newer.

## 2. Before and after

```go
// v2
pool := workerpool.New(100)
defer pool.Stop()

var wg sync.WaitGroup
worker := func(path string) {
	defer wg.Done()
	resize(path)
}
for i := 0; i < 4; i++ {
	if err := pool.AddWorker(worker); err != nil {
		return err
	}
}

wg.Add(len(paths))
for _, path := range paths {
	if err := pool.Delegate(path); err != nil {
		return err
	}
}
wg.Wait() // v2 couldn't wait for the queue itself
```

```go
// v3
pool := workerpool.New[string](100)
for range 4 {
	pool.AddWorker(resize) // resize is a func(path string)
}

for _, path := range paths {
	if err := pool.Delegate(ctx, path); err != nil {
		return err
	}
}
pool.Stop() // runs every queued job, then returns
```

The type parameter is the job's type, and `New` needs it spelled out: `workerpool.New[string](100)`.

## 3. Every call

| v2 | v3 |
| --- | --- |
| `New(queueLength) Pool` | `New[T](queueLength) *Pool[T]` |
| `type Pool interface { … }` | `type Pool[T any] struct { … }` |
| `pool.Delegate(args...) error` | `pool.Delegate(ctx, job) error` |
| `pool.AddWorker(fn) error` | `remove := pool.AddWorker(fn)` |
| `pool.RemoveWorker(fn) error` | `remove()` |
| `pool.WorkersNum() int` | `pool.WorkersNum() int` |
| `pool.Stop()` | `pool.Stop()` |
| — | `ErrStopped` |

A v2 job of several arguments becomes one value: `Delegate(n, out)` to a `func(n int, out chan<- int)` turns
into a struct, `Delegate(ctx, job{n, out})` to a `func(j job)`.

Keep the func `AddWorker` returns for each worker you'll want to remove. Calling it a second time does nothing.

## 4. Behaviour changes

- **Jobs and workers are typed.** A job whose type doesn't match the worker's func no longer compiles. In v2 it
  panicked in the worker's goroutine and took the program down. `AddWorker(nil)` panics with
  "workerpool: nil worker func"; v2 panicked with a nil pointer dereference.
- **Delegate doesn't need a worker.** v2 returned "there is no workers in pool" until a worker was added; v3
  queues the job, and it waits for one.
- **Delegate takes a context.** When the queue is full it waits for room until ctx is done, and returns ctx's
  error; v2 could only block. If ctx is already done, it returns ctx's error and queues nothing.
- **Stop is final, and waits.** v2's Stop removed the workers and left the queued jobs for workers added later.
  v3's Stop runs the queued jobs on the workers there are, waits for every worker to exit, and leaves the pool
  stopped. Jobs still queued when no worker is left are dropped, where v2 kept them for workers added later: `Delegate` returns `ErrStopped`, and calls that were waiting for room return too; `AddWorker` starts
  nothing. In v2, Delegate after Stop queued the job for workers that were gone and, once the queue was full,
  blocked forever. To pause a v3 pool, remove its workers and add new ones later; add them before `Stop`, which drops
  what no worker is left to run.
- **Stop and remove are safe to repeat.** In v2, a second `Stop`, or `RemoveWorker` after `Stop`, panicked
  with "close of closed channel".
- **remove takes out the worker it came from.** v2 looked workers up by their func, so a method value such as
  `pool.RemoveWorker(s.handle)` never matched the one passed to `AddWorker`, and an unknown func returned nil
  having removed nothing.
- **WorkersNum counts what is left.** It drops when a worker is removed and is 0 once `Stop` returns; v2 kept
  counting the workers `Stop` had removed.
- **No data race.** v2's `Delegate` and `WorkersNum` read the workers map while `AddWorker` and `RemoveWorker`
  wrote it.
- **A negative queue length** panics with "workerpool: negative queue length"; v2 panicked in `make`.
