package workerpool

import (
	"fmt"
	"reflect"
)

// Pool implements worker pool
type Pool interface {
	// Delegate job to a workers
	// will block if channel is full, you might want to wrap it with goroutine to avoid it
	// will panic if called after Stop()
	Delegate(args ...interface{})
	// Start given number of workers that will take jobs from a queue
	Start(maxWorkers int, fn interface{}) error
	// Stop all workers
	Stop()
}

type pool struct {
	queue         chan []reflect.Value
	isQueueClosed bool
}

func (p *pool) Delegate(args ...interface{}) {
	p.queue <- buildQueueValue(args)
}

func (p *pool) Start(maxWorkers int, fn interface{}) error {
	if maxWorkers < 1 {
		return fmt.Errorf("Invalid number of workers: %d", maxWorkers)
	}

	if reflect.TypeOf(fn).Kind() != reflect.Func {
		return fmt.Errorf("%s is not a reflect.Func", reflect.TypeOf(fn))
	}

	if p.isQueueClosed {
		return fmt.Errorf("Can not start already stopped worker")
	}

	for i := 1; i <= maxWorkers; i++ {
		h := reflect.ValueOf(fn)

		go func() {
			for args := range p.queue {
				h.Call(args)
			}
		}()
	}

	return nil
}

func (p *pool) Stop() {
	close(p.queue)
	p.isQueueClosed = true
}

func buildQueueValue(args []interface{}) []reflect.Value {
	reflectedArgs := make([]reflect.Value, 0)

	for _, arg := range args {
		reflectedArgs = append(reflectedArgs, reflect.ValueOf(arg))
	}

	return reflectedArgs
}

// New creates new worker pool with a given job queue length
func New(queueLength int) Pool {
	return &pool{
		queue: make(chan []reflect.Value, queueLength),
	}
}
