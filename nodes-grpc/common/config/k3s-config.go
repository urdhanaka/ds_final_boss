package config

import (
	"fmt"
	"os"
)

const (
	BasicCommand = `curl -sfL https://get.k3s.io | K3S_KUBECONFIG_MODE=644`
)

// check for k3s binary
//
// return true if binary exists and executable. Using
// default locations from k3s script
func checkK3SBinary() bool {
	isExistAndExecutable := false

    // default locations from k3s script
	locations := []string{"/usr/local/bin/k3s", "/opt/bin/k3s"}

	for _, location := range locations {
		if fileInfo, err := os.Stat(location); err == nil {
			fileMode := fileInfo.Mode()

			if fileMode&0111 == 0111 {
				isExistAndExecutable = true
			}
		}
	}

	return isExistAndExecutable
}

func StartCommand(isServer bool, token string, masterNodeIP string) string {
	var fullCommand string

	if checkK3SBinary() {
		fullCommand = fmt.Sprintf("%s INSTALL_K3S_SKIP_DOWNLOAD=true", BasicCommand)
	}

	if isServer {
		fullCommand = fmt.Sprintf(`%s INSTALL_K3S_EXEC="server" K3S_TOKEN=%s sh -s -`, BasicCommand, token)
	} else {
		fullCommand = fmt.Sprintf(`%s K3S_URL=https://%s:6443 K3S_TOKEN=%s sh -s -`, BasicCommand, masterNodeIP, token)
	}

	return fullCommand
}
