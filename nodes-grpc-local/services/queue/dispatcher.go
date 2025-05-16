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
	virtService  virtualization.VirtualizationInterface
	worker       chan struct{}
	jobQueue     *Queue
}

func NewDispatcher(
	jobQueue *Queue,
	virtService virtualization.VirtualizationInterface,
) *Dispatcher {
	return &Dispatcher{
		virtService: virtService,
		worker:      make(chan struct{}, MAX_WORKER_SIZE),
		jobQueue:    jobQueue,
	}
}

func (d *Dispatcher) AddJob(
	ctx context.Context,
	virtualizationRequest virtualization_model.CreateInstanceRequest,
) error {
	newJob := NewJob(virtualizationRequest)
        
    slog.Info("adding job to queue")

	isSuccess := d.jobQueue.TryAdd(ctx, newJob)
	if !isSuccess {
		return errors.New("could not add new job, try again later")
	}

	return nil
}

func (d *Dispatcher) Wait() {
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

			d.worker <- struct{}{}

			slog.Info("A worker is available, working on a job...")

			go func(j *Job) {
				defer func() { <-d.worker }()

				var err error

				for retry := 1; retry <= j.Retries; retry++ {
					err = d.virtService.CreateInstance(thisContext, job.Request)
					if err == nil {
						slog.Info("provisioning done")

						// testing the queue
						time.Sleep(time.Second * j.Backoff)

						return
					} else {
						slog.Error("provisioning error, trying in several seconds...")
						time.Sleep(time.Second * j.Backoff)
					}
				}

				slog.Error("provisioning error after trying",
					"error", err,
				)
			}(job)

		case <-thisContext.Done():
			slog.Info("cancelled")
		}
	}
}
