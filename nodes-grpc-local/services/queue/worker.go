package queue

import (
	"context"
	"encoding/json"
	"log/slog"
	libvirt_virtualization "nodes-grpc-local/services/virtualization/libvirt-virtualization"
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

func (w *Worker) DoWork() {
	ctx := context.Background()

	for {
		instanceRequest, err := w.queue.PopQueue(ctx)
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

		err = w.queue.Publish(ctx, instanceRequest.Name, string(jsonBytes))
		if err != nil {
			slog.Error("error creating instance",
				"error", err,
			)
			continue
		}
	}
}
