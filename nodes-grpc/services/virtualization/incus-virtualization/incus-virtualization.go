package incus_virtualization

import (
	"context"
	"fmt"
	"log/slog"
	virtualization_model "nodes-grpc/common/model/virtualization"
	"nodes-grpc/services/virtualization"
	"nodes-grpc/services/virtualization/embedded"
	virtualization_utils "nodes-grpc/services/virtualization/utils"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	incus "github.com/lxc/incus/client"
	"github.com/lxc/incus/shared/api"
)

const (
	BOTTOM_FORWARD_PORT = 49152
	UPPER_FORWARD_PORT  = 65535
)

type IncusVirtualization struct {
	incusConnection incus.InstanceServer
}

func NewIncusVirtualization(
	incusConnection incus.InstanceServer,
) virtualization.VirtualizationInterface {
	return &IncusVirtualization{
		incusConnection: incusConnection,
	}
}

// func (c *IncusVirtualization) Spawn(
// 	ctx context.Context,
// 	virtRequest virtualization_model.InstanceCreateRequest,
// ) error {
// 	slog.Info(fmt.Sprintf(
// 		"ID: %s, Spawn(): creating incus vm instance...", ctx.Value("contextId")),
// 	)
//
// 	profileFile, err := os.ReadFile("./common/templates/cloud-init-templates/user-cloud-init.yaml")
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
//
// 	newUuid := uuid.New().String()
// 	req := api.InstancesPost{
// 		InstancePut: api.InstancePut{
// 			Architecture: "amd64",
// 			Config: map[string]string{
// 				"security.secureboot":  "false",
// 				"cloud-init.user-data": string(profileFile),
// 				"limits.cpu":           "2",
// 				"limits.memory":        "2GiB",
// 			},
// 			Ephemeral: true,
// 			Profiles: []string{
// 				"default",
// 			},
// 		},
// 		Name: newUuid,
// 		Type: api.InstanceTypeVM,
// 		Source: api.InstanceSource{
// 			Type:  "image",
// 			Alias: "debian/bookworm/cloud",
// 			// Alias: "alpine/edge/cloud",
// 			// Properties: map[string]string{
// 			// 	"os":      "Debian",
// 			// 	"release": "bookworm",
// 			// 	"variant": "cloud",
// 			// },
// 			Server:   "https://images.linuxcontainers.org",
// 			Protocol: "simplestreams",
// 		},
// 		Start: true,
// 	}
//
// 	instanceOp, err := c.incusConnection.CreateInstance(req)
// 	if err != nil {
// 		return err
// 	}
//
// 	err = instanceOp.WaitContext(ctx)
// 	if err != nil {
// 		return err
// 	}
//
// 	// wait until cloud-init process is complete
// 	cloudExecReq := api.InstanceExecPost{
// 		Command: []string{
// 			"cloud-init", "status", "--wait",
// 		},
// 		WaitForWS: true,
// 	}
// 	cloudExecOp, err := c.incusConnection.ExecInstance(req.Name, cloudExecReq, nil)
// 	if err != nil {
// 		return err
// 	}
// 	err = cloudExecOp.WaitContext(ctx)
// 	if err != nil {
// 		slog.Error("cloudExec()",
// 			"error", err.Error(),
// 		)
//
// 		return err
// 	}
//
// 	execReq := api.InstanceExecPost{
// 		Command: []string{
// 			"bash", "-c", "curl -sfL https://get.k3s.io | sh -s -",
// 		},
// 		WaitForWS: true,
// 		Environment: map[string]string{
// 			"K3S_TOKEN": virtRequest.Token,
// 		},
// 		Cwd: "/root",
// 	}
// 	if virtRequest.IsMaster {
// 		execReq.Environment["INSTALL_K3S_EXEC"] = "server"
// 	} else {
// 		execReq.Environment["INSTALL_K3S_EXEC"] = "agent"
// 	}
// 	execOp, err := c.incusConnection.ExecInstance(req.Name, execReq, &incus.InstanceExecArgs{
// 		Stdout: os.Stdout,
// 	})
// 	if err != nil {
// 		return err
// 	}
//
// 	err = execOp.WaitContext(ctx)
// 	if err != nil {
// 		slog.Error("exec()",
// 			"error", err.Error(),
// 		)
//
// 		return err
// 	}
//
// 	slog.Info("Spawn(): spawning incus vm instance successful")
//
// 	return nil
// }

// stop the instance
//
// because the instance is set as ephemeral, we can just poweroff it
// and it will be deleted
func (c *IncusVirtualization) Stop(
	ctx context.Context,
	instanceIdentification virtualization_model.InstanceIdentification,
) error {
	slog.Info(fmt.Sprintf(
		"ID: %s, Stop(): stopping instance(s)...", ctx.Value("contextId")),
	)

	instanceName := instanceIdentification.InstanceID.String()

	poweroffExecReq := api.InstanceExecPost{
		Command: []string{
			"bash", "-c", "poweroff",
		},
		WaitForWS: true,
	}
	poweroffExecOp, err := c.incusConnection.ExecInstance(instanceName, poweroffExecReq, nil)
	if err != nil {
		return err
	}
	err = poweroffExecOp.WaitContext(ctx)
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf(
		"ID: %s, Stop(): instance %s successfully stopped", ctx.Value("contextId"), instanceName),
	)

	return nil
}

func (c *IncusVirtualization) List() error {
	return nil
}

func (c *IncusVirtualization) SpawnMaster(
	ctx context.Context,
	virtRequest virtualization_model.InstanceCreateRequest,
) (string, error) {
	slog.Info(fmt.Sprintf("%s: Spawn(): spawning incus master vm...", ctx.Value("contextId")))

	instanceName := uuid.New().String()
	err := c.spawnBase(ctx, instanceName, virtRequest)
	if err != nil {
		return "", err
	}

	slog.Info(fmt.Sprintf("%s: Spawn(): installing k3s...", ctx.Value("contextId")))

	// installing k3s
	k3sExecReq := api.InstanceExecPost{
		Command: []string{
			"bash", "-c", "curl -sfL https://get.k3s.io | sh -s -",
		},
		WaitForWS: true,
		Environment: map[string]string{
			"K3S_TOKEN":        virtRequest.Token,
			"INSTALL_K3S_EXEC": "server",
		},
	}
	k3sExecOp, err := c.incusConnection.ExecInstance(instanceName, k3sExecReq, &incus.InstanceExecArgs{
		Stdout: os.Stdout,
	})
	if err != nil {
		slog.Error(fmt.Sprintf("%s: Spawn(): could not install k3s", ctx.Value("contextId")),
			"error", err.Error(),
		)
		return "", err
	}
	err = k3sExecOp.WaitContext(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("%s: Spawn(): could not install k3s", ctx.Value("contextId")),
			"error", err.Error(),
		)
		return "", err
	}

	// installing helm
	slog.Info(fmt.Sprintf("%s: Spawn(): installing helm for dashboard...", ctx.Value("contextId")))
	helmExecReq := api.InstanceExecPost{
		Command: []string{
			"bash", "-c", "curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash",
		},
		WaitForWS: true,
	}
	helmExecOp, err := c.incusConnection.ExecInstance(instanceName, helmExecReq, &incus.InstanceExecArgs{
		Stdout: os.Stdout,
	})
	if err != nil {
		slog.Error(fmt.Sprintf("%s: Spawn(): could not install helm", ctx.Value("contextId")),
			"error", err.Error(),
		)
		return "", err
	}
	err = helmExecOp.WaitContext(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("%s: Spawn(): could not install helm", ctx.Value("contextId")),
			"error", err.Error(),
		)
		return "", err
	}

	// setting env var for helm
	slog.Info(fmt.Sprintf("%s: Spawn(): setting up env var for helm...", ctx.Value("contextId")))
	envExecReq := api.InstanceExecPost{
		Command: []string{
			"bash", "-c", "echo \"export KUBECONFIG=/etc/rancher/k3s/k3s.yaml\" >> ~/.bashrc",
		},
		WaitForWS: true,
	}
	envExecOp, err := c.incusConnection.ExecInstance(instanceName, envExecReq, nil)
	if err != nil {
		slog.Error(fmt.Sprintf("%s: Spawn(): could not set up env var", ctx.Value("contextId")),
			"error", err.Error(),
		)
		return "", err
	}
	err = envExecOp.WaitContext(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("%s: Spawn(): could not set up env var", ctx.Value("contextId")),
			"error", err.Error(),
		)
		return "", err
	}

	// setting up kubernetes dashboard admin-user
	slog.Info(fmt.Sprintf("%s: Spawn(): setting up kubernetes dashboard...", ctx.Value("contextId")))

	embeddedFolder := embedded.ReturnEmbedded()
	kubectlYaml, _ := embeddedFolder.ReadFile("create-admin-user.yaml")

	kubectlExecReq := api.InstanceExecPost{
		Command: []string{
			"bash", "-c", fmt.Sprintf("echo \"%s\" | k3s kubectl apply -f -", string(kubectlYaml)),
		},
		WaitForWS: true,
	}
	kubectlExecOp, err := c.incusConnection.ExecInstance(instanceName, kubectlExecReq, nil)
	if err != nil {
		slog.Error(fmt.Sprintf("%s: Spawn(): could not set up kubernetes dashboard", ctx.Value("contextId")),
			"error", err.Error(),
		)
		return "", err
	}
	err = kubectlExecOp.WaitContext(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("%s: Spawn(): could not set up kubernetes dashboard", ctx.Value("contextId")),
			"error", err.Error(),
		)
		return "", err
	}

	// get instance ip address
	networkAllocations, err := c.incusConnection.GetNetworkAllocations(false)
	if err != nil {
		slog.Error(fmt.Sprintf("%s: Spawn(): could not get instance network", ctx.Value("contextId")),
			"error", err.Error(),
		)
		return "", err
	}
	for _, allocation := range networkAllocations {
		// skip ipv6
		if strings.Contains(allocation.Address, ":") {
			continue
		}

		// get the instance used address
		if strings.Contains(allocation.UsedBy, instanceName) {
			fullAddress := allocation.Address
			splittedAddress := strings.Split(fullAddress, "/")

			return splittedAddress[0], nil
		}
	}

	slog.Info(fmt.Sprintf("%s: Spawn(): spawning incus master vm successful", ctx.Value("contextId")))

	return "", nil
}

func (c *IncusVirtualization) SpawnWorker(
	ctx context.Context,
	virtRequest virtualization_model.InstanceCreateRequest,
) error {
	slog.Info(fmt.Sprintf("%s: Spawn(): spawning incus worker vm...", ctx.Value("contextId")))

	instanceName := uuid.New().String()
	err := c.spawnBase(ctx, instanceName, virtRequest)
	if err != nil {
		return err
	}

	execReq := api.InstanceExecPost{
		Command: []string{
			"bash", "-c", "curl -sfL https://get.k3s.io | sh -s -",
		},
		WaitForWS: true,
		Environment: map[string]string{
			"K3S_TOKEN":        virtRequest.Token,
			"INSTALL_K3S_EXEC": "agent",
			"K3S_URL":          fmt.Sprintf("https://%s:6443", virtRequest.MasterIpAddress),
		},
	}
	execOp, err := c.incusConnection.ExecInstance(instanceName, execReq, &incus.InstanceExecArgs{
		Stdout: os.Stdout,
	})
	if err != nil {
		return err
	}
	err = execOp.WaitContext(ctx)
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("%s: Spawn(): spawning incus worker vm successful", ctx.Value("contextId")))

	return nil
}

func (c *IncusVirtualization) StopNode(
	ctx context.Context,
	instance virtualization_model.InstanceIdentification,
) error {
	return nil
}

func (c *IncusVirtualization) spawnBase(
	ctx context.Context,
	instanceName string,
	virtRequest virtualization_model.InstanceCreateRequest,
) error {
	profileFile, err := os.ReadFile("./common/templates/cloud-init-templates/user-cloud-init.yaml")
	if err != nil {
		return err
	}

	req := api.InstancesPost{
		InstancePut: api.InstancePut{
			Architecture: "amd64",
			Config: map[string]string{
				"security.secureboot":  "false",
				"cloud-init.user-data": string(profileFile),
				"limits.cpu":           fmt.Sprintf("%d", virtRequest.Cpu),
				"limits.memory":        fmt.Sprintf("%dGiB", virtRequest.Memory),
			},
			Ephemeral: true, // delete the instance when poweroff
			Profiles: []string{
				"default",
			},
		},
		Name: instanceName,
		Type: api.InstanceTypeVM,
		Source: api.InstanceSource{
			Type:  "image",
			Alias: "debian/bookworm/cloud",
			// Alias: "alpine/edge/cloud",
			// Properties: map[string]string{
			// 	"os":      "Debian",
			// 	"release": "bookworm",
			// 	"variant": "cloud",
			// },
			Server:   "https://images.linuxcontainers.org",
			Protocol: "simplestreams",
		},
		Start: true,
	}

	instanceOp, err := c.incusConnection.CreateInstance(req)
	if err != nil {
		return err
	}
	err = instanceOp.WaitContext(ctx)
	if err != nil {
		return err
	}

	// fucking vm-agent is not automatically running
	// wait for 20 seconds before executing any command
	time.Sleep(time.Second * 20)

	cloudExecReq := api.InstanceExecPost{
		Command: []string{
			"cloud-init", "status", "--wait",
		},
		WaitForWS: true,
	}
	cloudExecOp, err := c.incusConnection.ExecInstance(req.Name, cloudExecReq, nil)
	if err != nil {
		return err
	}
	err = cloudExecOp.WaitContext(ctx)
	if err != nil {
		return err
	}

	return nil
}

// TODO: proxy
// might not needed if we can use bridge network
func (c *IncusVirtualization) createProxy(
	ctx context.Context,
	instanceName string,
) error {
	var randomPort int
	var err error

	for true {
		randomPort, err = virtualization_utils.GetRandomPort(BOTTOM_FORWARD_PORT, UPPER_FORWARD_PORT)
		if err != nil {
			return err
		}

		isPortAvailable := virtualization_utils.IsPortAvailable(randomPort)
		if isPortAvailable {
			break
		}
	}

	// TODO:
	_, _, err = c.incusConnection.GetInstance(instanceName)
	if err != nil {
		return err
	}
	// append a new device config
	_ = map[string]map[string]string{
		"k3s-proxy": {
			"type": "proxy",
		},
	}

	return nil
}
