worker-pool
================
[![Build Status](https://travis-ci.org/vardius/worker-pool.svg?branch=master)](https://travis-ci.org/vardius/worker-pool)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/worker-pool)](https://goreportcard.com/report/github.com/vardius/worker-pool)
[![codecov](https://codecov.io/gh/vardius/worker-pool/branch/master/graph/badge.svg)](https://codecov.io/gh/vardius/worker-pool)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fvardius%2Fworker-pool.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fvardius%2Fworker-pool?ref=badge_shield)
[![](https://godoc.org/github.com/vardius/worker-pool?status.svg)](http://godoc.org/github.com/vardius/worker-pool)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/vardius/worker-pool/blob/master/LICENSE.md)

<img align="right" height="180px" src="https://github.com/vardius/gorouter/blob/master/website/src/static/img/logo.png?raw=true" alt="logo" />

Go simple async worker pool.

📖 ABOUT
==================================================
Contributors:

* [Rafał Lorenz](http://rafallorenz.com)

Want to contribute ? Feel free to send pull requests!

Have problems, bugs, feature ideas?
We are using the github [issue tracker](https://github.com/vardius/worker-pool/issues) to manage them.

## 📚 Documentation

For __examples__ **visit [godoc#pkg-examples](http://godoc.org/github.com/vardius/worker-pool#pkg-examples)**

For **GoDoc** reference, **visit [godoc.org](http://godoc.org/github.com/vardius/worker-pool)**

🚏 HOW TO USE
==================================================

## 🚅 Benchmark
**CPU: 3,3 GHz Intel Core i7**

**RAM: 16 GB 2133 MHz LPDDR3**

```bash
➜  worker-pool git:(master) ✗ go test -bench=. -cpu=4 -benchmem
goos: darwin
goarch: amd64
BenchmarkWorker1-4                	 3000000	       453 ns/op	      56 B/op	       3 allocs/op
BenchmarkWorker1Parallel-4        	 3000000	       506 ns/op	      48 B/op	       2 allocs/op
BenchmarkWorker100-4              	 3000000	       485 ns/op	      56 B/op	       3 allocs/op
BenchmarkWorker100Parallel-4      	 3000000	       444 ns/op	      48 B/op	       2 allocs/op
BenchmarkWorkerNumCPU-4           	 3000000	       467 ns/op	      56 B/op	       3 allocs/op
BenchmarkWorkerNumCPUParallel-4   	 3000000	       431 ns/op	      48 B/op	       2 allocs/op
PASS
ok  	worker-pool	11.570s
```

## 🏫 Basic example
```go
package main

import (
    "fmt"
    "sync"

    "github.com/vardius/worker-pool"
)

func main() {
	var wg sync.WaitGroup

	poolSize := 1
	jobsAmount := 3
	workersAmount := 2

	// create new pool
	pool := workerpool.New(poolSize)
	out := make(chan int, jobsAmount)

	pool.Start(workersAmount, func(i int) {
		defer wg.Done()
		out <- i
	})

	wg.Add(jobsAmount)

	for i := 0; i < jobsAmount; i++ {
		pool.Delegate(i)
	}

	go func() {
		// stop all workers after jobs are done
		wg.Wait()
		close(out)
		pool.Stop()
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
