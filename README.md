# 👨‍🔧 worker-pool

[![Test](https://github.com/vardius/worker-pool/actions/workflows/test.yml/badge.svg)](https://github.com/vardius/worker-pool/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/vardius/worker-pool/v3.svg)](https://pkg.go.dev/github.com/vardius/worker-pool/v3)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/worker-pool/v3)](https://goreportcard.com/report/github.com/vardius/worker-pool/v3)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE.md)

A worker pool for Go: one bounded queue of typed jobs, and the goroutines that run them.

- **Typed.** A `Pool[T]` queues values of type `T`, and a worker is a `func(T)`: no reflection, no `interface{}`.
- **Backpressure.** `Delegate` waits while the queue is full, and gives up when its context is done.
- **Graceful stop.** `Stop` runs the queued jobs on the workers left and waits for them. After it, `Delegate` returns `ErrStopped`.
- **Workers at any time.** `AddWorker` starts a worker and returns the func that removes it.
- **Lean.** Queueing a job into a queue with room and running it on a worker allocate nothing, and tests guard
  the allocation counts.

📖 **[Documentation](https://vardius.github.io/worker-pool/)** · [API reference](https://pkg.go.dev/github.com/vardius/worker-pool/v3) · [Examples](examples) · [Migrating from v2](MIGRATION.md)

## Install

```sh
go get github.com/vardius/worker-pool/v3
```

Requires Go 1.26 or newer.

## Usage

```go
pool := workerpool.New[string](100) // a queue of up to 100 paths
for range runtime.NumCPU() {
	pool.AddWorker(resize) // resize is a func(path string)
}

for _, path := range paths {
	if err := pool.Delegate(ctx, path); err != nil {
		return err
	}
}
pool.Stop() // runs what is still queued, then returns
```

## The queue

`Delegate` queues a job and returns at once while there is room. When the queue is full, it waits for a worker
to take a job, and returns ctx's error if ctx is done first, so a deadline bounds the wait:

```go
ctx, cancel := context.WithTimeout(r.Context(), 100*time.Millisecond)
defer cancel()
if err := pool.Delegate(ctx, job); err != nil {
	http.Error(w, "busy, try again later", http.StatusServiceUnavailable)
	return
}
```

Jobs wait in the queue until a worker is free, so they can be queued before any worker is added.

## Workers

Each worker is a goroutine that takes jobs one at a time and calls its func. Add workers at any time; each can
run a different func. `AddWorker` returns the func that removes that worker: it finishes the job it is running
and exits. One that is between jobs as you remove it may still take one more.

```go
remove := pool.AddWorker(send)
defer remove()
```

## Stopping

`Stop` stops the pool taking jobs, lets the workers run every job still queued, and returns once they have all
exited. With no worker left, the queued jobs are dropped. From then on `Delegate` returns `ErrStopped`. Calling `Stop` again is safe: it waits for the same.
To shut down faster, have the jobs watch a context, and cancel it before calling `Stop`.

## Benchmarks

Time and allocations per job, on an Apple M4 Max with Go 1.27.0, median of 5 runs. Both versions ran on the same
harness, with an `int` job and workers that do nothing. A queue with room never fills; the others hold 1024 jobs,
so `Delegate` runs at the workers' pace. With 16 callers at once, the time is wall time per call across all of
them, so it measures throughput. Run v3's own suite with `go test -bench .`.

| Scenario | v2 | v3 |
| --- | ---: | ---: |
| Delegate, queue has room | 32.6 ns · 2 allocs | **11.4 ns · 0 allocs** |
| Delegate, queue has room, 16 callers at once | 86.0 ns · 2 allocs | **35.4 ns · 0 allocs** |
| A job through 1 worker | 216 ns · 2 allocs | **38.8 ns · 0 allocs** |
| A job through 16 workers | 412 ns · 2 allocs | **110 ns · 0 allocs** |
| A job through 1 worker, 16 callers at once | 372 ns · 2 allocs | **86.9 ns · 0 allocs** |
| A job through 16 workers, 16 callers at once | 258 ns · 2 allocs | **55.4 ns · 0 allocs** |

v2 built a slice of `reflect.Value`s for every job and called the worker through reflection. v3 sends the job
on a typed channel. With jobs that do nothing, one caller can't keep 16 workers busy: they wait and wake for
every job, which costs more than the job itself.

## Upgrading from v2

```go
// v2
pool := workerpool.New(100)
worker := func(path string) { resize(path) }
_ = pool.AddWorker(worker)
_ = pool.Delegate(path)
_ = pool.RemoveWorker(worker)

// v3
pool := workerpool.New[string](100)
remove := pool.AddWorker(resize)
_ = pool.Delegate(ctx, path)
remove()
```

[MIGRATION.md](MIGRATION.md) maps every v2 call to its v3 form and lists the behaviour changes.

## Contributing

Bug reports and pull requests are welcome: see [CONTRIBUTING.md](.github/CONTRIBUTING.md).

## License

MIT, see [LICENSE.md](LICENSE.md).
