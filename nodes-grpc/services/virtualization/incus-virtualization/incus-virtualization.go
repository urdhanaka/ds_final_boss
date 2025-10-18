package incus_virtualization

import (
	"context"
	"fmt"
	"log/slog"
	virtualization_model "nodes-grpc/services/model/virtualization-model"
	"nodes-grpc/services/virtualization"
	"nodes-grpc/services/virtualization/embedded"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	incus "github.com/lxc/incus/client"
	"github.com/lxc/incus/shared/api"
)

const (
	// bridge name
	BRIDGE_NAME = "k3s-bridge0"

	// TODO: get rid of this
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

// stop the instance
//
// because the instance is set as ephemeral, we can just turn it off
// and it will be deleted
func (c *IncusVirtualization) Stop(
	ctx context.Context,
	instanceIdentification virtualization_model.InstanceIdentification,
) error {
	instanceName := instanceIdentification.InstanceID.String()

	incusSlogFunction(instanceName, "stopping instance(s)...", nil)

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
	err = poweroffExecOp.Wait()
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

func (c *IncusVirtualization) CreateMaster(
	ctx context.Context,
	virtRequest virtualization_model.NodeCreateRequest,
) (string, error) {
	instanceName := uuid.New().String()

	incusSlogFunction(instanceName, "starting master instance...", nil)
	err := c.createBase(ctx, instanceName, virtRequest)
	if err != nil {
		return "", err
	}

	// installing k3s
	incusSlogFunction(
		instanceName,
		"installing k3s...",
		nil,
	)
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
		incusSlogFunction(instanceName, "could not install k3s", err)

		return "", err
	}
	err = k3sExecOp.Wait()
	if err != nil {
		incusSlogFunction(instanceName, "could not install k3s", err)

		return "", err
	}

	// installing helm
	incusSlogFunction(instanceName, "installing helm for dashboard...", nil)
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
		incusSlogFunction(instanceName, "could not install helm", err)

		return "", err
	}
	err = helmExecOp.Wait()
	if err != nil {
		incusSlogFunction(instanceName, "could not install helm", err)

		return "", err
	}

	// setting env var for helm
	incusSlogFunction(instanceName, "setting env var for helm and kubectl...", nil)
	envExecReq := api.InstanceExecPost{
		Command: []string{
			"bash", "-c", "echo \"export KUBECONFIG=/etc/rancher/k3s/k3s.yaml\" >> ~/.bashrc",
		},
		WaitForWS: true,
	}
	envExecOp, err := c.incusConnection.ExecInstance(instanceName, envExecReq, nil)
	if err != nil {
		incusSlogFunction(instanceName, "could not set up env var", err)

		return "", err
	}
	err = envExecOp.Wait()
	if err != nil {
		incusSlogFunction(instanceName, "could not set up env var", err)

		return "", err
	}

	// setting up kubernetes dashboard admin-user
	incusSlogFunction(instanceName, "setting up kubernetes dashboard...", nil)

	embeddedFolder := embedded.ReturnEmbedded()
	kubectlYaml, _ := embeddedFolder.ReadFile("files/create-admin-user.yaml")

	kubectlExecReq := api.InstanceExecPost{
		Command: []string{
			"bash", "-c", fmt.Sprintf("echo \"%s\" | k3s kubectl apply -f -", string(kubectlYaml)),
		},
		WaitForWS: true,
	}
	kubectlExecOp, err := c.incusConnection.ExecInstance(instanceName, kubectlExecReq, nil)
	if err != nil {
		incusSlogFunction(instanceName, "could not set up kubernetes dashboard", err)

		return "", err
	}
	err = kubectlExecOp.Wait()
	if err != nil {
		incusSlogFunction(instanceName, "could not set up kubernetes dashboard", err)

		return "", err
	}

	// get instance ip address
	// ugly hack that use ip route and then grep the default line
	ipRouteExec := exec.Command("incus", "exec", instanceName, "--", "ip", "route", "|", "grep", "default")
	out, err := ipRouteExec.Output()
	if err != nil {
		incusSlogFunction(instanceName, "could not get instance ip address", err)

		return "", err
	}

	stringArr := strings.Fields(string(out))
	for _, str := range stringArr {
		// check if current str is an ip address
		// easy way: check if string contains dot (.)
		if strings.Contains(str, ".") {
			ipOctets := strings.Split(str, ".")

			if len(ipOctets) != 4 {
				continue
			}

			// if it is, split by the dot
			// and check the last string
			// if the value is not 1 (gateway), return it
			if ipOctets[len(ipOctets)-1] != "1" {
				return str, nil
			}
		}
	}

	incusSlogFunction(instanceName, "spawning master instance successful", nil)

	return "", nil
}

func (c *IncusVirtualization) CreateWorker(
	ctx context.Context,
	virtRequest virtualization_model.NodeCreateRequest,
) error {
	instanceName := uuid.New().String()

	incusSlogFunction(instanceName, "spawning worker instance", nil)

	err := c.createBase(ctx, instanceName, virtRequest)
	if err != nil {
		return err
	}

	incusSlogFunction(instanceName, "installing k3s", nil)

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
		incusSlogFunction(instanceName, "error spawning worker instance", err)

		return err
	}
	err = execOp.Wait()
	if err != nil {
		incusSlogFunction(instanceName, "error spawning worker instance", err)

		return err
	}

	incusSlogFunction(instanceName, "spawning worker instance success", nil)

	return nil
}

func (c *IncusVirtualization) StopNode(
	ctx context.Context,
	instance virtualization_model.InstanceIdentification,
) error {
	return nil
}

func (c *IncusVirtualization) createBase(
	ctx context.Context,
	instanceName string,
	virtRequest virtualization_model.NodeCreateRequest,
) error {
	embedded := embedded.ReturnEmbedded()
	profileFile, err := embedded.ReadFile("files/user-cloud-init.yaml")
	if err != nil {
		incusSlogFunction(instanceName, "error reading cloud-init file", err)

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
			Devices: map[string]map[string]string{
				"eth0": {
					"name":    "eth0",
					"nictype": "bridged",
					"parent":  BRIDGE_NAME,
					"type":    "nic",
				},
				"root": {
					"path": "/",
					"pool": "default",
					"type": "disk",
				},
			},
		},
		Name: instanceName,
		Type: api.InstanceTypeVM,
		Source: api.InstanceSource{
			Type:     "image",
			Alias:    "debian/bookworm/cloud",
			Server:   "https://images.linuxcontainers.org",
			Protocol: "simplestreams",
		},
		Start: true,
	}
	instanceOp, err := c.incusConnection.CreateInstance(req)
	if err != nil {
		incusSlogFunction(instanceName, "error creating base instance", err)

		return err
	}
	err = instanceOp.Wait()
	if err != nil {
		incusSlogFunction(instanceName, "error creating base instance", err)

		return err
	}

	time.Sleep(time.Second * 20)

	// start cloud init operation
	incusSlogFunction(instanceName, "waiting for cloud-init operation", nil)

	cloudExecReq := api.InstanceExecPost{
		Command: []string{
			"cloud-init", "status", "--wait",
		},
		WaitForWS: true,
	}
	cloudExecOp, err := c.incusConnection.ExecInstance(req.Name, cloudExecReq, nil)
	if err != nil {
		incusSlogFunction(instanceName, "cloud-init operation error", err)

		return err
	}
	err = cloudExecOp.Wait()
	if err != nil {
		incusSlogFunction(instanceName, "starting cloud-init operation error ", err)

		return err
	}

	return nil
}
