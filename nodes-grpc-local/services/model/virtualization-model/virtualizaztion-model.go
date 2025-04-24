package virtualization_model

type CreateInstanceRequest struct {
	// is the instance the master?
	IsMaster bool

	// token for k3s
	Token string

	// if spawning worker
	// need to know the master IP address
	MasterIpAddress string

	// vm requirements
	Cpu     int64
	Memory  int64
	Storage int64
}
