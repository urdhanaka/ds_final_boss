package virtualization_model

type CreateInstanceRequest struct {
	// cluster name
	ClusterName string `json:"cluster_name"`

	// instance name
	Name string `json:"name"`

	// token for k3s
	Token string `json:"token,omitempty"`

	// if spawning worker
	// need to know the master IP address
	MasterIpAddress string `json:"master_ip_address,omitempty"`

	// vm requirements
	Cpu     int `json:"cpu"`
	Memory  int `json:"memory"`
	Storage int `json:"storage"`

	// is the instance the master?
	IsMaster bool `json:"isMaster"`

	// main server ip address
	MainServerIpAddress string `json:"main_server_ip_address"`

	// cluster UUID
	ClusterId string `json:"cluster_id"`
}

type DeleteInstanceRequest struct {
	Name string `json:"name"`
}

type Instance struct {
	Name   string
	Status bool
}

type PidQemuGuestAgent struct {
	Return Pid `json:"return"`
}

type Pid struct {
	PID int `json:"pid"`
}

type ExecStatusQemuGuestAgent struct {
	Return ExecStatus `json:"return"`
}

type ExecStatus struct {
	Exited   bool `json:"exited,omitempty"`
	ExitCode int  `json:"exitcode,omitempty"`

	// this struct contains string of base-64, make sure
	// to convert to string if string data type is needed
	OutData string `json:"out-data,omitempty"`
}

type VirtCreateInstanceResponse struct {
	KubeconfigContents []byte `json:"kubeconfig_contents,omitempty"`
	MasterIpAddress    string `json:"master_ip_address,omitempty"`
	DashboardToken     string `json:"dashboard_token,omitempty"`
	Message            string `json:"message,omitempty"`
	CreationStatus     bool   `json:"creation_status,omitempty"`
}
