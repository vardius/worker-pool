package workerpool

import (
	"context"
	"fmt"
	"reflect"
)

// Pool implements worker pool
type Pool interface {
	Delegate(args ...interface{})
	Start(maxWorkers int, fn interface{}) error
	Stop()
}

type pool struct {
	ctx    context.Context
	cancel context.CancelFunc
	queue  chan []reflect.Value
}

// Delegate job to a workers
// will block if channel is full, you might want to wrap it with goroutine to avoid it
// will panic if called after Stop()
func (p *pool) Delegate(args ...interface{}) {
	p.queue <- buildQueueValue(args)
}

// Start given number of workers that will take jobs from a queue
func (p *pool) Start(maxWorkers int, fn interface{}) error {
	if maxWorkers < 1 {
		return fmt.Errorf("Invalid number of workers: %d", maxWorkers)
	}

	if reflect.TypeOf(fn).Kind() != reflect.Func {
		return fmt.Errorf("%s is not a reflect.Func", reflect.TypeOf(fn))
	}

	if err := p.ctx.Err(); err != nil {
		return err
	}

	for i := 1; i <= maxWorkers; i++ {
		h := reflect.ValueOf(fn)

		go func() {
			for {
				select {
				case args, ok := <-p.queue:
					if !ok {
						return
					}

					h.Call(args)
				case <-p.ctx.Done():
					return
				}
			}
		}()
	}

	return nil
}

// Stop all workers
func (p *pool) Stop() {
	defer close(p.queue)
	p.cancel()
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
	ctx, cancel := context.WithCancel(context.Background())

	return &pool{
		ctx:    ctx,
		cancel: cancel,
		queue:  make(chan []reflect.Value, queueLength),
	}
}
