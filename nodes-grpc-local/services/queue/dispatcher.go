package queue

import (
	"context"
	"errors"
	"log/slog"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"
	"nodes-grpc-local/services/virtualization"
	"sync"
	"time"
)

type Dispatcher struct {
	dispatcherWg sync.WaitGroup
	worker       chan struct{}
	jobQueue     *Queue
	virtService  virtualization.VirtualizationInterface
}

func NewDispatcher(
	jobQueue *Queue,
	virtService virtualization.VirtualizationInterface,
) *Dispatcher {
	return &Dispatcher{
		worker:      make(chan struct{}, MAX_WORKER_SIZE),
		jobQueue:    jobQueue,
		virtService: virtService,
	}
}

func (d *Dispatcher) AddJob(
	ctx context.Context,
	virtualizationRequest virtualization_model.CreateInstanceRequest,
) error {
	newJob := NewJob(virtualizationRequest)

	isSuccess := d.jobQueue.TryAdd(ctx, newJob)
	if !isSuccess {
		return errors.New("could not add new job, try again later")
	}

	return nil
}

func (d *Dispatcher) Wait(ctx context.Context) {
	d.dispatcherWg.Wait()
}

func (d *Dispatcher) Start() {
	defer d.dispatcherWg.Done()
	d.dispatcherWg.Add(1)

	thisContext := context.Background()

	for {
		select {
		case job, ok := <-d.jobQueue.jobs:
			if !ok {
				slog.Info("worker Shutting down")
				return
			}

			slog.Info("received job, working...")

			d.worker <- struct{}{}

			go func(j *Job) {
				defer d.dispatcherWg.Done()
				defer func() { <-d.worker }()

				for retry := 1; retry <= j.Retries; retry++ {
					err := d.virtService.CreateInstance(thisContext, job.Request)
					if err == nil {
						slog.Info("provisioning done")
						return
					} else {
						slog.Error("provisioning error, trying in several seconds...")
						time.Sleep(time.Second * j.Backoff)
					}
				}
			}(job)

		case <-thisContext.Done():
			slog.Info("cancelled")
		}
	}
}
