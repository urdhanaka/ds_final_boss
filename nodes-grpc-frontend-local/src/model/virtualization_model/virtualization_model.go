package virtualization_model

type CreateClusterRequest struct {
	Name    string `json:"name"`
	VCPU    string `json:"vcpu"`
	Memory  string `json:"memory"`
	Storage string `json:"storage"`
	Master  int    `json:"master_node"`
	Worker  int    `json:"worker_node"`
}
