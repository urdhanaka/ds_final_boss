package queue

import (
	"errors"
	"sync"
	"time"
)

type Worker struct {
	localWg *sync.WaitGroup
}

func NewWorker(localWg *sync.WaitGroup) Worker {
	return Worker{
		localWg: localWg,
	}
}

func (w Worker) DoWork(jobQueue <-chan Job) error {
	defer w.localWg.Done()

	w.localWg.Add(1)

	for {
		select {
		case _, ok := <-jobQueue:
			if !ok {
				return errors.New("queue is closed")
			}

		case <-time.After(time.Second * DEFAULT_JOB_TIMEOUT):
			return errors.New("timeout")
		}
	}
}
