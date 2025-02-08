package utils

import (
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}

	return hostname
}

func GetNodeIP() (string, error) {
	nodeInterfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, nodeInterface := range nodeInterfaces {
		interfaceName := nodeInterface.Name

		// skip loopback interface
		if strings.Index(interfaceName, "lo") == 0 {
			continue
		}

		// handling ethernet interface
		// common ethernet interface name: "en", "eth"
		if strings.Index(nodeInterface.Name, "en") == 0 ||
			strings.Index(nodeInterface.Name, "eth") == 0 {
			addrs, err := nodeInterface.Addrs()
			if err != nil {
				return "", err
			}

			for _, addr := range addrs {
				// handle ipv4 only for now
				if addr.Network() == "ip+net" && strings.Contains(addr.String(), ".") {
					ipOnly := strings.Split(addr.String(), "/")[0]

					return ipOnly, nil
				}
			}
		}

		// handling wireless interface
		if strings.Index(nodeInterface.Name, "wl") == 0 {
			addrs, err := nodeInterface.Addrs()
			if err != nil {
				return "", err
			}

			for _, addr := range addrs {
				// handle ipv4 only for now
				if addr.Network() == "ip+net" && strings.Contains(addr.String(), ".") {
					ipOnly := strings.Split(addr.String(), "/")[0]

					return ipOnly, nil
				}
			}
		}
	}

	return "", nil
}

func RandomString(length int) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, length)
	for i := range b {
		b[i] = letters[random.Intn(len(letters))]
	}

	return string(b)
}
