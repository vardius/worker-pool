/*
Package workerpool provides simple async workers

Basic example:
	package main

	import (
		"fmt"
		"sync"

		"github.com/vardius/worker-pool"
	)

	func main() {
		var wg sync.WaitGroup

		// create new pool
		pool := workerpool.New(poolSize)
		out := make(chan int, jobsAmount)

		go func() {
			// stop all workers after jobs are done
			defer pool.Stop()
			wg.Wait()
			close(out)
		}()

		pool.Start(2, func(i int) {
			defer wg.Done()
			out <- i
		})

		wg.Add(workersAmount)

		for i := 0; i < jobsAmount; i++ {
			pool.Delegate(i)
		}

		sum := 0
		for n := range out {
			sum += n
		}

		fmt.Println(sum)
	}
*/
package workerpool
