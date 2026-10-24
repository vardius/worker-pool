# Changelog

## v3.0.0 (2026-10-24)

A generic rewrite with no reflection. See [MIGRATION.md](MIGRATION.md) for the before and after of every call.

### Added

- A type parameter on `Pool` and `New`: a job is a `T` and a worker is a `func(T)`, checked by the compiler.
- `AddWorker` returns the func that removes that worker, in place of `RemoveWorker`.
- A context on `Delegate`: when the queue is full, it waits for room until the context is done.
- `ErrStopped`, which `Delegate` returns once the pool is stopped.
- Runnable examples in `examples/`, each with a test: hashing files on every CPU, and an HTTP server that
  answers 503 when its queue stays full and finishes the queued jobs on shutdown.

### Changed

- Module path is `github.com/vardius/worker-pool/v3`; Go 1.26 or newer is required.
- `Pool` is a struct, `*Pool[T]`, instead of an interface.
- `Stop` runs the queued jobs on the workers left, waits for every worker to exit, and leaves the pool stopped.
  Jobs still queued when no worker is left are dropped. It can be called more than once.
- `Delegate` queues jobs when there are no workers; they wait for one to be added.
- A removed worker finishes its job and exits.
- `WorkersNum` is 0 once `Stop` returns.
- Faster, with no allocations. Queueing a job went from 33 ns and 2 allocations to 11 ns and none, and a job
  through one worker from 216 ns to 39 ns. Allocation counts are guarded by tests.

### Fixed

- A data race between `Delegate` or `WorkersNum` and `AddWorker` or `RemoveWorker`.
- A second `Stop`, or removing a worker after `Stop`, panicked with "close of closed channel".
- `Delegate` after `Stop` queued jobs for workers that were gone and blocked forever once the queue was full.
- Removing a worker added as a method value removed nothing, and removing an unknown func returned nil.

### Removed

- `RemoveWorker`, replaced by the func `AddWorker` returns.
- The `Pool` interface, and jobs of several arguments: pass one value, a struct if need be.
