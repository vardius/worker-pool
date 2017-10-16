Vardius - worker-pool
================
[![Build Status](https://travis-ci.org/vardius/worker-pool.svg?branch=master)](https://travis-ci.org/vardius/worker-pool)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/worker-pool)](https://goreportcard.com/report/github.com/vardius/worker-pool)
[![codecov](https://codecov.io/gh/vardius/worker-pool/branch/master/graph/badge.svg)](https://codecov.io/gh/vardius/worker-pool)
[![](https://godoc.org/github.com/vardius/worker-pool?status.svg)](http://godoc.org/github.com/vardius/worker-pool)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/vardius/worker-pool/blob/master/LICENSE.md)

Go simple async worker pool.

ABOUT
==================================================
Contributors:

* [Rafał Lorenz](http://rafallorenz.com)

Want to contribute ? Feel free to send pull requests!

Have problems, bugs, feature ideas?
We are using the github [issue tracker](https://github.com/vardius/worker-pool/issues) to manage them.

HOW TO USE
==================================================

1. [GoDoc](http://godoc.org/github.com/vardius/worker-pool)

## Benchmark
**CPU: 3,3 GHz Intel Core i7
RAM: 16 GB 2133 MHz LPDDR3**
```bash
➜  worker-pool git:(master) ✗ go test -bench=. -cpu=4
BenchmarkWorker-4        1000000              1853 ns/op
PASS
```

## Basic example
```go
package main

import (
    "fmt"
    "sync"

    "github.com/vardius/worker-pool"
)

func main() {
    var wg sync.WaitGroup

    poolSize: 1
    jobsAmount: 3
    workersAmount: 2

    // create new pool
    pool := workerpool.New(poolSize)
    out := make(chan int, jobsAmount)

    pool.Start(workersAmount, func(i int) {
        defer wg.Done()
        out <- i
    })

    wg.Add(workersAmount)

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
}
```

License
-------

This package is released under the MIT license. See the complete license in the package:

[LICENSE](LICENSE.md)
