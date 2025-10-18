package services

import (
	"nodes-grpc/services/virtualization"
	incus_virtualization "nodes-grpc/services/virtualization/incus-virtualization"
)

type Connection struct {
	VirtualizationService virtualization.VirtualizationInterface
}

func NewConnection() *Connection {
	// connection
	incusConnection := virtualization.InitIncusConnection()

	// services
	incusService := incus_virtualization.NewIncusVirtualization(incusConnection)

	return &Connection{
		VirtualizationService: incusService,
	}
}
