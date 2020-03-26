👨‍🔧 worker-pool
================
[![Build Status](https://travis-ci.org/vardius/worker-pool.svg?branch=master)](https://travis-ci.org/vardius/worker-pool)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/worker-pool)](https://goreportcard.com/report/github.com/vardius/worker-pool)
[![codecov](https://codecov.io/gh/vardius/worker-pool/branch/master/graph/badge.svg)](https://codecov.io/gh/vardius/worker-pool)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fvardius%2Fworker-pool.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fvardius%2Fworker-pool?ref=badge_shield)
[![](https://godoc.org/github.com/vardius/worker-pool?status.svg)](https://pkg.go.dev/github.com/vardius/worker-pool)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/vardius/worker-pool/blob/master/LICENSE.md)

<img align="right" height="180px" src="https://github.com/vardius/gorouter/blob/master/website/src/static/img/logo.png?raw=true" alt="logo" />

Go simple async worker pool.

📖 ABOUT
==================================================

<img align="left" src="https://brandur.org/assets/images/go-worker-pool/worker-pool.svg" href="https://brandur.org/go-worker-pool" />

Worker pool is a software design pattern for achieving concurrency of task execution. Maintains multiple workers waiting for tasks to be allocated for concurrent execution. By maintaining a pool of workers, the model increases performance and avoids latency in execution. The number of available workers might be tuned to the computing resources available.

You can read more about worker pools in Go [here](https://brandur.org/go-worker-pool).

Contributors:

* [Rafał Lorenz](http://rafallorenz.com)

Want to contribute ? Feel free to send pull requests!

Have problems, bugs, feature ideas?
We are using the github [issue tracker](https://github.com/vardius/worker-pool/issues) to manage them.

## 📚 Documentation

For __examples__ **visit [godoc#pkg-examples](http://godoc.org/github.com/vardius/worker-pool#pkg-examples)**

For **GoDoc** reference, **visit [pkg.go.dev](https://pkg.go.dev/github.com/vardius/worker-pool)**

🚏 HOW TO USE
==================================================

## 🚅 Benchmark
**CPU: 3,3 GHz Intel Core i7**

**RAM: 16 GB 2133 MHz LPDDR3**

```bash
➜  worker-pool git:(master) ✗ go test -bench=. -cpu=4 -benchmem
goos: darwin
goarch: amd64
pkg: github.com/vardius/worker-pool/v2
BenchmarkWorker1-4                	 3944299	       284 ns/op	      56 B/op	       3 allocs/op
BenchmarkWorker1Parallel-4        	 7394715	       138 ns/op	      48 B/op	       2 allocs/op
BenchmarkWorker100-4              	 1657569	       693 ns/op	      56 B/op	       3 allocs/op
BenchmarkWorker100Parallel-4      	 3673483	       368 ns/op	      48 B/op	       2 allocs/op
BenchmarkWorkerNumCPU-4           	 2590293	       445 ns/op	      56 B/op	       3 allocs/op
BenchmarkWorkerNumCPUParallel-4   	 3591553	       298 ns/op	      48 B/op	       2 allocs/op
PASS
ok  	github.com/vardius/worker-pool/v2	9.511s
```

## 🏫 Basic example
```go
package main

import (
    "fmt"
    "sync"

    "github.com/vardius/worker-pool/v2"
)

func main() {
	var wg sync.WaitGroup

	poolSize := 1
	jobsAmount := 3
	workersAmount := 2

	// create new pool
	pool := workerpool.New(poolSize)
	out := make(chan int, jobsAmount)
	worker := func(i int) {
        defer wg.Done()
        out <- i
    }

	for i := 1; i <= workersAmount; i++ {
		if err := pool.AddWorker(worker); err != nil {
			panic(err)
		}
	}

	wg.Add(jobsAmount)

	for i := 0; i < jobsAmount; i++ {
		if err := pool.Delegate(i); err != nil {
			panic(err)
		}
	}

	go func() {
		// stop all workers after jobs are done
		wg.Wait()
		close(out)
		pool.Stop() // stop removes all workers from pool, to resume work add them again
	}()

	sum := 0
	for n := range out {
		sum += n
	}

	fmt.Println(sum)
	// Output:
	// 3
}
```

📜 [License](LICENSE.md)
-------

This package is released under the MIT license. See the complete license in the package

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fvardius%2Fworker-pool.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fvardius%2Fworker-pool?ref=badge_large)
