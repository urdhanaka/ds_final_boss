package virtualization

import "net"

type IncusVirtualization struct {
	SockConnection net.Conn
}

func NewIncusVirtualization() *IncusVirtualization {
	return &IncusVirtualization{
		SockConnection: initIncusConnection(),
	}
}

func (v IncusVirtualization) GetAllInstances() {
}
