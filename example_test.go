package workerpool_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vardius/worker-pool/v3"
)

func Example() {
	var mu sync.Mutex
	total := 0

	pool := workerpool.New[int](10)
	for range 3 {
		pool.AddWorker(func(n int) {
			mu.Lock()
			defer mu.Unlock()
			total += n
		})
	}

	for n := 1; n <= 100; n++ {
		if err := pool.Delegate(context.Background(), n); err != nil {
			fmt.Println(err)
			return
		}
	}
	pool.Stop() // returns once every queued job has run

	fmt.Println(total)
	// Output: 5050
}

func ExamplePool_Delegate() {
	pool := workerpool.New[string](1)
	defer pool.Stop()

	// No worker yet, so the first job fills the queue.
	fmt.Println(pool.Delegate(context.Background(), "first"))

	// The queue is full: Delegate waits for room until ctx is done.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	fmt.Println(pool.Delegate(ctx, "second"))
	// Output:
	// <nil>
	// context deadline exceeded
}

func ExamplePool_AddWorker() {
	pool := workerpool.New[string](0)
	defer pool.Stop()

	greet := func(name string) { fmt.Println("hello,", name) }
	remove := pool.AddWorker(greet)
	pool.AddWorker(greet) // the same func again: a second worker, with its own remove

	fmt.Println(pool.WorkersNum())
	remove()
	fmt.Println(pool.WorkersNum())
	// Output:
	// 2
	// 1
}

func ExamplePool_Stop() {
	pool := workerpool.New[string](3)
	for _, job := range []string{"a", "b", "c"} {
		_ = pool.Delegate(context.Background(), job)
	}

	// Jobs queued before the worker was added still run: Stop waits for them.
	pool.AddWorker(func(job string) { fmt.Println("ran", job) })
	pool.Stop()

	fmt.Println(pool.Delegate(context.Background(), "d"))
	// Output:
	// ran a
	// ran b
	// ran c
	// workerpool: pool stopped
}
