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
	// IsMaster    bool
	// Hostname    string

	Cpu         int64
	Memory      int64
	MemorySize  string // B, MB, GB, etc..
	Storage     int64
	StorageSize string // B, MB, GB, etc..
}

type VirtualizationCreateResponse struct {
	IpAddress string
}
