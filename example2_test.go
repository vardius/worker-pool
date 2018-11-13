package workerpool_test

import (
	"fmt"
	"strconv"
	"sync"

	workerpool "github.com/vardius/worker-pool"
)

func Example_second() {
	poolSize := 2
	jobsAmount := 8
	workersAmount := 3

	var wg sync.WaitGroup
	wg.Add(jobsAmount)

	// allocate queue
	pool := workerpool.New(poolSize)

	// moc arg
	argx := make([]string, jobsAmount)
	for j := 0; j < jobsAmount; j++ {
		argx[j] = "_" + strconv.Itoa(j) + "_"
	}

	// assign job
	for i := 0; i < jobsAmount; i++ {
		go func(i int) {
			pool.Delegate(argx[i])
		}(i)
	}

	// start worker
	pool.Start(workersAmount, func(s string) {
		defer wg.Done()
		defer fmt.Println("job " + s + " is done !")
		fmt.Println("job " + s + " is running ..")
	})

	// clean up
	wg.Wait()
	pool.Stop()

	// fmt.Println("# hi: ok?")
	// Output:
	// # sq: let-me-check
}
