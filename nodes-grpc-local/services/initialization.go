package services

import (
	"nodes-grpc-local/services/virtualization"
	libvirt_virtualization "nodes-grpc-local/services/virtualization/libvirt-virtualization"
)

type Connection struct {
	VirtualizationService virtualization.VirtualizationInterface
}

func NewConnection() *Connection {
	libvirtConnection := libvirt_virtualization.InitLibvirtConnection()
	libvirtService := libvirt_virtualization.NewLibvirtVirtualization(libvirtConnection)

	return &Connection{
		VirtualizationService: libvirtService,
	}
}
