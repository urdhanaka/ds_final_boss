package services

import (
	"nodes-grpc-local/services/queue"
	"nodes-grpc-local/services/virtualization"
	libvirt_virtualization "nodes-grpc-local/services/virtualization/libvirt-virtualization"
)

type InitStruct struct {
	VirtualizationService virtualization.VirtualizationInterface
	DispatcherService     *queue.Dispatcher
}

func NewInitStruct() *InitStruct {
	libvirtConnection := libvirt_virtualization.InitLibvirtConnection()
	libvirtService := libvirt_virtualization.NewLibvirtVirtualization(libvirtConnection)

	queueStruct := queue.NewQueue()
	dispatcherService := queue.NewDispatcher(queueStruct, libvirtService)

	return &InitStruct{
		VirtualizationService: libvirtService,
		DispatcherService:     dispatcherService,
	}
}
