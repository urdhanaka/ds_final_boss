package services

import (
	"nodes-grpc/services/db"
	"nodes-grpc/services/virtualization"
)

type Connection struct {
	DockerService  *virtualization.DockerVirtualization
	IncusService   *virtualization.IncusVirtualization
	LibvirtService *virtualization.LibvirtVirtualization
}

func NewConnection() *Connection {
	// connection
	dbConnection := db.InitDB()
	// dockerConnection := virtualization.InitDockerConnection()
	incusConnection := virtualization.InitIncusConnection()
	libvirtConnection := virtualization.InitLibvirtConnection()

	// services
	dbService := db.NewDBConnection(dbConnection)
	// dockerService := virtualization.NewDockerVirtualization(dockerConnection, dbService)
	incusService := virtualization.NewIncusVirtualization(incusConnection, dbService)
	libvirtService := virtualization.NewLibvirtService(libvirtConnection, dbService)

	return &Connection{
		// DockerService:  dockerService,
		IncusService:   incusService,
		LibvirtService: libvirtService,
	}
}
