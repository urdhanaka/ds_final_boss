package main

import (
	"log"
	"os"

	incus "github.com/lxc/incus/client"
	"github.com/lxc/incus/shared/api"
)

func main() {
	c, err := incus.ConnectIncusUnix("/run/incus/unix.socket", nil)
	if err != nil {
		log.Fatalf(err.Error())
	}

	profileFile, err := os.ReadFile("./cloud-init.yaml")
	if err != nil {
		log.Fatalf(err.Error())
	}

	cloudProfile := api.ProfilesPost{
		ProfilePut: api.ProfilePut{
			Config: map[string]string{
				"cloud-init.user-data": string(profileFile),
			},
		},
		Name: "cloud-init-profile",
	}

	err = c.CreateProfile(cloudProfile)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// newUuid := uuid.New().String()
	// req := api.InstancesPost{
	// 	InstancePut: api.InstancePut{
	// 		Architecture: "amd64",
	// 		Config: map[string]string{
	// 			"security.secureboot": "false",
	// 		},
	// 	},
	// 	Name: newUuid,
	// 	Type: api.InstanceTypeVM,
	// 	Source: api.InstanceSource{
	// 		Type:  "image",
	// 		Alias: "debian/bookworm/cloud",
	// 		// Alias: "alpine/edge/cloud",
	// 		// Properties: map[string]string{
	// 		// 	"os":      "Debian",
	// 		// 	"release": "bookworm",
	// 		// 	"variant": "cloud",
	// 		// },
	// 		Server:   "https://images.linuxcontainers.org",
	// 		Protocol: "simplestreams",
	// 	},
	// 	Start: true,
	// }
	//
	// op, err := c.CreateInstance(req)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }
	// err = op.Wait()
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }
}
