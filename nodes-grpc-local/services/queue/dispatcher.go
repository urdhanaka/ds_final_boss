package queue

import (
	"context"
	"sync"
)

type Dispatcher struct {
	workerPool chan struct{}
	jobQueue   chan *Job
	globalWg   sync.WaitGroup
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

func (d *Dispatcher) Start(ctx context.Context) {
	d.globalWg.Add(1)
}
