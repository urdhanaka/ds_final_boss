package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"
	"nodes-grpc-local/services/virtualization"
	"time"

	"github.com/valkey-io/valkey-glide/go/api"
	"github.com/valkey-io/valkey-glide/go/api/options"
)

type Worker struct {
	valkeyClient          api.GlideClientCommands
	virtualizationService virtualization.VirtualizationInterface
}

func NewWorker(
	valkeyClient api.GlideClientCommands,
	virtualizationService virtualization.VirtualizationInterface,
) *Worker {
	return &Worker{
		valkeyClient,
		virtualizationService,
	}
}

func (w *Worker) DoWork(ctx context.Context) {
	thisWorkerContext, cancel := context.WithCancel(ctx)
	thisWorkerId := generateRandom(8)

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

		slog.Info(fmt.Sprintf("worker %s working...", thisWorkerId))

		err = w.virtualizationService.CreateInstance(thisWorkerContext, instanceRequest)
		if err != nil {
			slog.Error("error creating instance",
				"error", err,
			)
			cancel()
		}

		// testing for worker
		// imitating a long process
		time.Sleep(15 * time.Second)
	}
}
