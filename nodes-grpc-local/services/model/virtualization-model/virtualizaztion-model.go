package virtualization_model

type CreateInstanceRequest struct {
	// instance name
	Name string `json:"name"`

	// token for k3s
	Token string `json:"token,omitempty"`

	// if spawning worker
	// need to know the master IP address
	MasterIpAddress string `json:"master_ip_address,omitempty"`

	// vm requirements
	Cpu     int64 `json:"cpu"`
	Memory  int64 `json:"memory"`
	Storage int64 `json:"storage"`

	// is the instance the master?
	IsMaster bool `json:"isMaster"`
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
