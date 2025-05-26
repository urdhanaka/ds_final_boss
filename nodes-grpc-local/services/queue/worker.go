package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"
	libvirt_virtualization "nodes-grpc-local/services/virtualization/libvirt-virtualization"

	"github.com/valkey-io/valkey-glide/go/api"
	"github.com/valkey-io/valkey-glide/go/api/options"
)

type Worker struct {
	valkeyClient api.GlideClientCommands
	virtService  *libvirt_virtualization.LibvirtVirtualization
	id           int
}

func NewWorker(
	valkeyClient api.GlideClientCommands,
	virtService *libvirt_virtualization.LibvirtVirtualization,
	id int,
) *Worker {
	return &Worker{
		valkeyClient,
		virtService,
		id,
	}
}

func (w *Worker) DoWork(ctx context.Context) {
	thisWorkerContext, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		job, err := w.valkeyClient.BLMove(
			VALKEY_MAIN_QUEUE,
			VALKEY_PROCESSING_QUEUE,
			options.Left,
			options.Right,
			api.CreateNilFloat64Result().Value(),
		)
		if err != nil {
			slog.Error("processing queue error",
				"error", err,
			)
			cancel()
		}

		var instanceRequest virtualization_model.CreateInstanceRequest
		err = json.Unmarshal([]byte(job.Value()), &instanceRequest)
		if err != nil {
			slog.Error("json unmarshal error",
				"error", err,
			)
			cancel()
		}

		slog.Info(fmt.Sprintf("worker-%d working...", w.id))

		err = w.virtService.CreateInstance(thisWorkerContext, instanceRequest)
		if err != nil {
			slog.Error("error creating instance",
				"error", err,
			)
			cancel()
		}
	}
}
