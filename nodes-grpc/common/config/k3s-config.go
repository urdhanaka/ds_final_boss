package config

import (
	"fmt"
	"os"
)

const (
	// get the k3s script and set the kubeconfig file to 644 mode (readable by all group)
	BasicCommand = `curl -sfL https://get.k3s.io | K3S_KUBECONFIG_MODE=644`
)

// check for k3s binary
//
// return true if binary exists and executable. Using
// default locations from k3s script (/usr/local/bin/k3s, /opt/bin/k3s)
func checkK3SBinary() bool {
	isExistAndExecutable := false

	// default locations from k3s script
	locations := []string{"/usr/local/bin/k3s", "/opt/bin/k3s"}

	for _, location := range locations {
		if fileInfo, err := os.Stat(location); err == nil {
			// check if it's executable by every group
			fileMode := fileInfo.Mode()
			if fileMode&0111 == 0111 {
				isExistAndExecutable = true
			}
		}
	}

	return isExistAndExecutable
}

func GetStartCommand(isServer bool, token string, masterNodeIP string) string {
	var fullCommand string

	// check if k3s binary exists
	// if yes, skip downloading the k3s binary
	if checkK3SBinary() {
		fullCommand = fmt.Sprintf("%s INSTALL_K3S_SKIP_DOWNLOAD=true", BasicCommand)
	}

	if isServer {
		// start as a server with token
		fullCommand = fmt.Sprintf(`%s INSTALL_K3S_EXEC="server" K3S_TOKEN=%s sh -s -`, BasicCommand, token)
	} else {
		// start as a worker with token
		fullCommand = fmt.Sprintf(`%s K3S_URL=https://%s:6443 K3S_TOKEN=%s sh -s -`, BasicCommand, masterNodeIP, token)
	}

	return fullCommand
}
