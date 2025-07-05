package libvirt_virtualization

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
	ErrData string `json:"err-data,omitempty"`
}
