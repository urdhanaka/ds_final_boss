package services

import (
	"nodes-grpc-local/services/queue"
	libvirt_virtualization "nodes-grpc-local/services/virtualization/libvirt-virtualization"
	"nodes-grpc-local/services/websocket"
)

type InitStruct struct {
	VirtualizationService *libvirt_virtualization.LibvirtVirtualization
	WebsocketService      *websocket.Websocket
	QueueService          *queue.Queue
}

func NewInitStruct() *InitStruct {
	redisClient := queue.InitRedisConnection()

	queueStruct := queue.NewQueue(redisClient)

	websocketConnection := websocket.NewWebsocket()

	libvirtConnection := libvirt_virtualization.InitLibvirtConnection()
	libvirtService := libvirt_virtualization.NewLibvirtVirtualization(libvirtConnection, websocketConnection)

	return &InitStruct{
		VirtualizationService: libvirtService,
		QueueService:          queueStruct,
		WebsocketService:      websocketConnection,
	}
}
