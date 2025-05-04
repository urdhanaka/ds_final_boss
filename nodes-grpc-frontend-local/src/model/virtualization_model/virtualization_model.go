package virtualization_model

type CreateClusterRequest struct {
	VCPU    string `json:"vcpu"`
	Memory  string `json:"memory"`
	Storage string `json:"storage"`
}
