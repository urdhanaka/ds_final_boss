package model

type CreateClusterRequest struct {
	Name          string `json:"name"`
	VCPU          string `json:"vcpu"`
	Memory        string `json:"memory"`
	Storage       string `json:"storage"`
	NumberOfNodes int    `json:"number_of_nodes"`
	Master        int    `json:"master_node"`
	Worker        int    `json:"worker_node"`
}

type VirtCreateInstanceResponse struct {
	MasterIpAddress string `json:"master_ip_address"`
	Status          bool   `json:"status"`
	DashboardToken  string `json:"dashboard_token,omitempty"`
}
