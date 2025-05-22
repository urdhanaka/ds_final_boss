package services

import (
	"nodes-grpc-local/services/queue"
	"nodes-grpc-local/services/virtualization"
	libvirt_virtualization "nodes-grpc-local/services/virtualization/libvirt-virtualization"
	"nodes-grpc-local/services/websocket"
)

type InitStruct struct {
	VirtualizationService virtualization.VirtualizationInterface
	DispatcherService     *queue.Dispatcher
	WebsocketService      *websocket.Websocket
	QueueService          *queue.Queue
}

func NewInitStruct() *InitStruct {
	valkeyClient := queue.InitValkeyConnection()

	websocketConnection := websocket.NewWebsocket()

	libvirtConnection := libvirt_virtualization.InitLibvirtConnection()
	libvirtService := libvirt_virtualization.NewLibvirtVirtualization(libvirtConnection, websocketConnection)

	queueStruct := queue.NewQueue(valkeyClient)
	dispatcherService := queue.NewDispatcher(queueStruct, libvirtService)

	return &InitStruct{
		VirtualizationService: libvirtService,
		DispatcherService:     dispatcherService,
		QueueService:          queueStruct,
		WebsocketService:      websocketConnection,
	}
}
