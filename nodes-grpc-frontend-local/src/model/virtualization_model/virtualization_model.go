package virtualization_model

type CreateClusterRequest struct {
	Name     string `json:"name"`
	VCPU     string `json:"vcpu"`
	Memory   string `json:"memory"`
	Storage  string `json:"storage"`
	NodeSize string `json:"node_size"`
	Master   int    `json:"master_node"`
	Worker   int    `json:"worker_node"`
}

type VirtCreateInstanceResponse struct {
	MasterIpAddress string `json:"master_ip_address"`
	Status          bool   `json:"status"`
	DashboardToken  string `json:"dashboard_token,omitempty"`
}
