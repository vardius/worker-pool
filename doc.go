/*
Package workerpool provides simple async workers

Basic example:
	package main

	import (
		"fmt"

		"github.com/vardius/worker-pool"
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
*/
package workerpool
