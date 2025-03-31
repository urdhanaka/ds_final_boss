package virtualization_model

type VirtualizationCreateRequest struct {
	ClusterName string
	NodesNumber int

	Cpu         int64
	Memory      int64
	MemorySize  string // B, MB, GB, etc..
	Storage     int64
	StorageSize string // B, MB, GB, etc..
}

type NodeCreateRequest struct {
	// is the instance the master?
	IsMaster bool

	// token for k3s
	Token string

	// vm requirements
	Cpu     int64
	Memory  int64
	Storage int64
}

type VirtualizationCreateResponse struct {
	IpAddress string
}
