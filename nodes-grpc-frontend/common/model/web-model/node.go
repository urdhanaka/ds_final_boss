package web_model

type RegisterNodeRequest struct {
	Hostname  string `json:"hostname"`
	IpAddress string `json:"ip_address"`
	GrpcPort  string `json:"grpc_port"`
}

type GetNodesResponse struct {
	Hostname  string `json:"hostname"`
	IpAddress string `json:"ip_address"`
	GrpcPort  string `json:"grpc_port"`
	Status    string `json:"status"`
}

type CreateClusterRequest struct {
	Token string `json:"token"`
}

type NodeRequirement struct {
	Memory string `json:"memory"`
	VCPU   string `json:"vcpu"`
}
