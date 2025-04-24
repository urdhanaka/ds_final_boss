package services

import "nodes-grpc-local/services/virtualization"

type Connection struct {
	VirtualizationService virtualization.VirtualizationInterface
}

func NewConnection() *Connection {
	// incusConnection := virtualization.InitIncusConnection()
	libvirtConnection := virtualization.InitLibvirtConnection()

	return &Connection{
		VirtualizationService: libvirtConnection,
	}
}
