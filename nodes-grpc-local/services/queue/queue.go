package queue

import (
	"context"
	"encoding/json"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"

	"github.com/valkey-io/valkey-glide/go/api"
)

type Queue struct {
	valkeyClient api.GlideClientCommands
	jobs         chan *Job
}

func NewQueue(
	valkeyClient api.GlideClientCommands,
) *Queue {
	return &Queue{
		valkeyClient: valkeyClient,
		jobs:         make(chan *Job, 30),
	}
}

// add new job to queue
func (s *Queue) AddToQueue(
	ctx context.Context,
	instanceRequest virtualization_model.CreateInstanceRequest,
) error {
	requestString, err := json.Marshal(instanceRequest)
	if err != nil {
		return err
	}

	_, err = s.valkeyClient.RPush(VALKEY_MAIN_QUEUE, []string{string(requestString)})
	if err != nil {
		return err
	}

	return nil
}

// TryAdd will try to add the job to the queue
//
// return true if job is successfully added,
// return false otherwise
func (s *Queue) TryAdd(ctx context.Context, job *Job) bool {
	select {
	case s.jobs <- job:
		return true
	default:
		return false
	}
}
