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

## Basic example
```go
package main

import (
    "fmt"

    workerPool "github.com/vardius/worker-pool"
)

func main() {
    pool := workerPool.New(2)

    out := make(chan int, 3)
    defer close(out)

    pool.Start(2, func(i int) {
        out <- i
    })

    pool.Delegate(1)
    pool.Delegate(2)
    pool.Delegate(3)

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
