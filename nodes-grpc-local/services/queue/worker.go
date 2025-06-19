package queue

import (
	"context"
	"encoding/json"
	"log/slog"
	libvirt_virtualization "nodes-grpc-local/services/virtualization/libvirt-virtualization"
)

const (
	MESSAGE_PUBLISH_RETRY_COUNT = 3
)

type Worker struct {
	queue       *Queue
	virtService *libvirt_virtualization.LibvirtVirtualization
}

func NewWorker(
	queue *Queue,
	virtService *libvirt_virtualization.LibvirtVirtualization,
) *Worker {
	return &Worker{
		queue,
		virtService,
	}
}

func (w *Worker) DoSpawnWork() {
	ctx := context.Background()

	for {
		instanceRequest, err := w.queue.PopSpawnQueue(ctx)
		if err != nil {
			slog.Error("error creating instance",
				"error", err,
			)
			continue
		}

		res, err := w.virtService.CreateInstance(ctx, instanceRequest)
		if err != nil {
			slog.Error("error creating instance",
				"error", err,
			)
			continue
		}

		jsonBytes, err := json.Marshal(res)
		if err != nil {
			slog.Error("error creating instance",
				"error", err,
			)
			continue
		}

		err = w.publishResult(ctx, instanceRequest.Name, string(jsonBytes))
		if err != nil {
			slog.Error("error publishing result, retrying...",
				"error", err,
			)

			// start retrying here
			for currentTry := 1; currentTry <= MESSAGE_PUBLISH_RETRY_COUNT; currentTry++ {
				err = w.publishResult(ctx, instanceRequest.Name, string(jsonBytes))
			}
		}
	}
}

func (w *Worker) DoDeleteWork() {
	ctx := context.Background()

	for {
		deleteRequest, err := w.queue.PopDeleteQueue(ctx)
		if err != nil {
			slog.Error("error deleting instance",
				"error", err,
			)
			continue
		}

		w.virtService.DeleteInstance(deleteRequest.Name)
	}
}

func (w *Worker) publishResult(
	ctx context.Context,
	instanceName string,
	message any,
) error {
	return w.queue.Publish(ctx, instanceName, message)
}
