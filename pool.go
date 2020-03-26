package workerpool

import (
	"fmt"
	"reflect"
	"sync"
)

// Pool implements worker pool
type Pool interface {
	// Delegate job to a workers
	// will block if channel is full, you might want to wrap it with goroutine to avoid it
	// will panic if called after Stop()
	Delegate(args ...interface{})

	// AddWorker adds worker to the pool
	AddWorker(fn interface{}) error
	// RemoveWorker removes worker from the pool
	RemoveWorker(fn interface{}) error

	// WorkersNum returns number of workers in the pool
	WorkersNum() int
	// Stop all workers
	Stop()
}

type pool struct {
	queue         chan []reflect.Value
	isQueueClosed bool
	workers       []reflect.Value
	mtx           sync.RWMutex
}

func (p *pool) Delegate(args ...interface{}) {
	p.queue <- buildQueueValue(args)
}

func (p *pool) AddWorker(fn interface{}) error {
	if err := isValidHandler(fn); err != nil {
		return err
	}

	if p.isQueueClosed {
		return fmt.Errorf("can not add new worker to already stopped pool")
	}

	worker := reflect.ValueOf(fn)

	go func() {
		for args := range p.queue {
			worker.Call(args)
		}
	}()

	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.workers = append(p.workers, worker)

	return nil
}

func (p *pool) RemoveWorker(fn interface{}) error {
	if err := isValidHandler(fn); err != nil {
		return err
	}

	rv := reflect.ValueOf(fn)

	p.mtx.Lock()
	defer p.mtx.Unlock()

	for i, worker := range p.workers {
		if worker == rv {
			p.workers = append(p.workers[:i], p.workers[i+1:]...)
		}
	}

	return nil
}

func (p *pool) WorkersNum() int {
	return len(p.workers)
}

func (p *pool) Stop() {
	close(p.queue)
	p.isQueueClosed = true
}

func isValidHandler(fn interface{}) error {
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		return fmt.Errorf("%s is not a reflect.Func", reflect.TypeOf(fn))
	}

	return nil
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
		queue:   make(chan []reflect.Value, queueLength),
		workers: make([]reflect.Value, 0),
	}
}
