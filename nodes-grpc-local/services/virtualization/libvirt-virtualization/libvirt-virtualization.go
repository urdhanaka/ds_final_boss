package libvirt_virtualization

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"
	"nodes-grpc-local/services/websocket"
	"os"
	"os/exec"
	"strings"
	"time"

	"libvirt.org/go/libvirt"
	"libvirt.org/go/libvirtxml"
)

// these hardcoded IP is for testing only
const (
	MASTER_NODE_IP = "192.168.122.49"
	WORKER_NODE_IP = "192.168.122.50"
)

const (
	// timeout in second when waiting the cloud init operations
	CLOUD_INIT_TIMEOUT = 60

	SHUTDOWN_RETRIES = 3
)

type LibvirtVirtualization struct {
	libvirtConnection   *libvirt.Connect
	websocketConnection *websocket.Websocket
}

func NewLibvirtVirtualization(
	libvirtConnection *libvirt.Connect,
	websocketConnection *websocket.Websocket,
) *LibvirtVirtualization {
	return &LibvirtVirtualization{
		libvirtConnection:   libvirtConnection,
		websocketConnection: websocketConnection,
	}
}

// wrapper for create master and create worker
// this function will be used in queue
func (c *LibvirtVirtualization) CreateInstance(
	ctx context.Context,
	virtRequest virtualization_model.CreateInstanceRequest,
) error {
	if virtRequest.IsMaster {
		return c.createMaster(ctx, virtRequest)
	} else {
		return c.createWorker(ctx, virtRequest)
	}
}

func (c *LibvirtVirtualization) createMaster(
	ctx context.Context,
	virtRequest virtualization_model.CreateInstanceRequest,
) error {
	thisInstanceName := virtRequest.Name

	c.websocketConnection.AddMap(thisInstanceName)

	slog.Info(fmt.Sprintf("master node name is %s", thisInstanceName))
	slogFunction(thisInstanceName, "creating master instance", nil)
	c.websocketConnection.AddLogToMap(thisInstanceName, "creating master instance")

	// create network
	c.websocketConnection.AddLogToMap(thisInstanceName, "creating node network")
	slogFunction(thisInstanceName, "creating node network", nil)
	err := createNetworkMaster()
	if err != nil {
		slogFunction(thisInstanceName, "could not create node network", err)
		c.deleteInstance(thisInstanceName)

		return err
	}

	// create cloudinit configuration
	c.websocketConnection.AddLogToMap(thisInstanceName, "creating cloud-init configuration")
	slogFunction(thisInstanceName, "creating node cloud-init configuration", nil)
	err = createCloudInitMaster(thisInstanceName)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node cloud-init configuration", err)
		c.deleteInstance(thisInstanceName)

		return err
	}

	// creating the image
	c.websocketConnection.AddLogToMap(thisInstanceName, "creating base image for instance")
	slogFunction(thisInstanceName, "creating base image for instance", nil)
	err = copyImage(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node image", err)
		c.deleteInstance(thisInstanceName)

		return err
	}

	// configure the EFI
	c.websocketConnection.AddLogToMap(thisInstanceName, "creating instance EFI file")
	slogFunction(thisInstanceName, "creating instance EFI file", nil)
	err = copyEfi(thisInstanceName)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node EFI", err)
		c.deleteInstance(thisInstanceName)

		return err
	}

	// base xml for the vm
	c.websocketConnection.AddLogToMap(thisInstanceName, "creating instance base xml")
	slogFunction(thisInstanceName, "creating instance base xml", nil)
	domainXmlConfig, err := createBase(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node base xml", err)
		c.deleteInstance(thisInstanceName)

		return err
	}

	// spawning
	c.websocketConnection.AddLogToMap(thisInstanceName, "spawning node")
	slogFunction(thisInstanceName, "spawning node", nil)
	dom, err := c.libvirtConnection.DomainCreateXML(domainXmlConfig, libvirt.DomainCreateFlags(0))
	if err != nil {
		slogFunction(thisInstanceName, "could not spawn node", err)
		c.deleteInstance(thisInstanceName)

		return err
	}

	slogFunction(thisInstanceName, "waiting until the vm is ready..", nil)
	waitCloudInitCmd := `{"execute":"guest-exec","arguments":{"path":"/bin/bash","arg":["-c", "cloud-init status --wait"],"capture-output":true}}`
	time.Sleep(15 * time.Second)
	cloudinitPid, err := dom.QemuAgentCommand(waitCloudInitCmd, libvirt.DOMAIN_QEMU_AGENT_COMMAND_BLOCK, 0)
	if err != nil {
		slogFunction(thisInstanceName, "could not spawn node", err)
		c.deleteInstance(thisInstanceName)

		return err
	}

	// get PID
	pidStruct := virtualization_model.PidQemuGuestAgent{}
	err = json.Unmarshal([]byte(cloudinitPid), &pidStruct)
	if err != nil {
		slogFunction(thisInstanceName, "could not spawn node", err)
		c.deleteInstance(thisInstanceName)

		return err
	}

	// check PID status
	cloudinitStatusCmd := fmt.Sprintf(`{"execute":"guest-exec-status","arguments":{"pid":%d}}`, pidStruct.Return.PID)
	for {
		cloudinitStatus, err := dom.QemuAgentCommand(cloudinitStatusCmd, libvirt.DOMAIN_QEMU_AGENT_COMMAND_BLOCK, 0)
		if err != nil {
			slogFunction(thisInstanceName, "could not spawn node", err)
			c.deleteInstance(thisInstanceName)

			return err
		}

		execStatusStruct := virtualization_model.ExecStatusQemuGuestAgent{}
		err = json.Unmarshal([]byte(cloudinitStatus), &execStatusStruct)
		if err != nil {
			slogFunction(thisInstanceName, "could not spawn node", err)
			c.deleteInstance(thisInstanceName)

			return err
		}

		if execStatusStruct.Return.Exited {
			break
		}

		time.Sleep(10 * time.Second)
	}

	c.websocketConnection.AddLogToMap(thisInstanceName, "done")
	return nil
}

func (c *LibvirtVirtualization) createWorker(
	ctx context.Context,
	virtRequest virtualization_model.CreateInstanceRequest,
) error {
	thisInstanceName := virtRequest.Name

	c.websocketConnection.AddMap(thisInstanceName)

	slog.Info(fmt.Sprintf("worker node name is %s", thisInstanceName))
	slogFunction(thisInstanceName, "creating worker instance", nil)
	c.websocketConnection.AddLogToMap(thisInstanceName, "creating worker instance")

	// create network
	c.websocketConnection.AddLogToMap(thisInstanceName, "creating instance network")
	slogFunction(thisInstanceName, "creating instance network", nil)
	err := createNetworkWorker()
	if err != nil {
		slogFunction(thisInstanceName, "could not create node network", err)
		c.deleteInstance(thisInstanceName)

		return err
	}

	// create cloudinit configuration
	c.websocketConnection.AddLogToMap(thisInstanceName, "creating cloud-init configuration")
	slogFunction(thisInstanceName, "creating node cloud-init configuration", nil)
	err = createCloudInitWorker(thisInstanceName)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node cloud-init configuration", err)
		c.deleteInstance(thisInstanceName)

		return err
	}

	// creating the image
	c.websocketConnection.AddLogToMap(thisInstanceName, "creating base image for instance")
	slogFunction(thisInstanceName, "creating node image", nil)
	err = copyImage(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node image", err)
		c.deleteInstance(thisInstanceName)

		return err
	}

	// configure the EFI
	c.websocketConnection.AddLogToMap(thisInstanceName, "creating instance EFI file")
	slogFunction(thisInstanceName, "creating node EFI", nil)
	err = copyEfi(thisInstanceName)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node EFI", err)
		c.deleteInstance(thisInstanceName)

		return err
	}

	// base xml for the vm
	c.websocketConnection.AddLogToMap(thisInstanceName, "creating instance base xml")
	slogFunction(thisInstanceName, "creating node base xml", nil)
	domainXmlConfig, err := createBase(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node base xml", err)
		c.deleteInstance(thisInstanceName)

		return err
	}

	// spawning
	c.websocketConnection.AddLogToMap(thisInstanceName, "spawning node")
	slogFunction(thisInstanceName, "spawning node", nil)
	dom, err := c.libvirtConnection.DomainCreateXML(domainXmlConfig, libvirt.DOMAIN_NONE)
	if err != nil {
		slogFunction(thisInstanceName, "could not spawn node", err)
		c.deleteInstance(thisInstanceName)

		return err
	}
	slogFunction(thisInstanceName, "waiting until the vm is ready..", nil)
	time.Sleep(15 * time.Second)
	waitCloudInitCmd := `{"execute":"guest-exec","arguments":{"path":"/bin/bash","arg":["-c", "cloud-init status --wait"],"capture-output":true}}`
	_, err = dom.QemuAgentCommand(waitCloudInitCmd, libvirt.DOMAIN_QEMU_AGENT_COMMAND_BLOCK, 0)
	if err != nil {
		slogFunction(thisInstanceName, "could not spawn node", err)
		c.deleteInstance(thisInstanceName)

		return err
	}

	c.websocketConnection.AddLogToMap(thisInstanceName, "done")
	return nil
}

func createBase(
	instanceName string,
	instanceConfig virtualization_model.CreateInstanceRequest,
) (string, error) {
	instanceStorage := POOL_DIR + "/" + instanceName + ".qcow2"
	seedFile := POOL_DIR + "/" + instanceName + ".iso"
	serialSocket := SERIAL_DIR + "/" + instanceName + ".sock"
	consoleSocket := CONSOLE_DIR + "/" + instanceName + ".sock"

	domConfig := &libvirtxml.Domain{
		Type: "kvm",
		Name: instanceName,
		Metadata: &libvirtxml.DomainMetadata{
			XML: `
<libosinfo:libosinfo xmlns:libosinfo="http://libosinfo.org/xmlns/libvirt/domain/1.0">
  <libosinfo:os id="http://ubuntu.com/ubuntu/24.10"/>
</libosinfo:libosinfo>`,
		},
		Memory: &libvirtxml.DomainMemory{
			Value: uint(instanceConfig.Memory),
			Unit:  "GB",
		},
		VCPU: &libvirtxml.DomainVCPU{
			Value: uint(instanceConfig.Cpu),
		},
		OS: &libvirtxml.DomainOS{
			Firmware: "efi",
			Type: &libvirtxml.DomainOSType{
				Arch:    "x86_64",
				Machine: "pc-q35-9.2",
				Type:    "hvm",
			},
			FirmwareInfo: &libvirtxml.DomainOSFirmwareInfo{
				Features: []libvirtxml.DomainOSFirmwareFeature{
					{
						Enabled: "no",
						Name:    "enrolled-keys",
					},
					{
						Enabled: "no",
						Name:    "secure-boot",
					},
				},
			},
			Loader: &libvirtxml.DomainLoader{
				Readonly: "yes",
				Type:     "pflash",
				Format:   "raw",
				Path:     LOADER_LOCAL, // NOTE: works on local only
				// Path: LOADER,
			},
			NVRam: &libvirtxml.DomainNVRam{
				NVRam:    fmt.Sprintf("/var/lib/libvirt/qemu/nvram/%s_VARS.fd", instanceName),
				Template: NVRAM_TEMPLATE_LOCAL, // NOTE: only works in local
				// Template:       NVRAM_TEMPLATE, // NOTE: directory according to ubuntu 24.04
				TemplateFormat: "raw",
				Format:         "raw",
			},
			BootDevices: []libvirtxml.DomainBootDevice{
				{
					Dev: "hd",
				},
			},
		},
		Features: &libvirtxml.DomainFeatureList{
			ACPI: &libvirtxml.DomainFeature{},
			APIC: &libvirtxml.DomainFeatureAPIC{},
			VMPort: &libvirtxml.DomainFeatureState{
				State: "off",
			},
		},
		CPU: &libvirtxml.DomainCPU{
			Mode: "host-passthrough",
		},
		OnPoweroff: "destroy",
		OnCrash:    "destroy",
		Devices: &libvirtxml.DomainDeviceList{
			Emulator: "/usr/bin/qemu-system-x86_64",
			Disks: []libvirtxml.DomainDisk{
				{
					Device: "disk",
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "qcow2",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: instanceStorage,
						},
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "vda",
						Bus: "virtio",
					},
				},
				{
					Device: "cdrom",
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "raw",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: seedFile,
						},
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "sda",
						Bus: "sata",
					},
					ReadOnly: &libvirtxml.DomainDiskReadOnly{},
				},
			},
			Interfaces: []libvirtxml.DomainInterface{
				{
					Source: &libvirtxml.DomainInterfaceSource{
						Network: &libvirtxml.DomainInterfaceSourceNetwork{
							Network: BRIDGE_NAME,
							Bridge:  BRIDGE_NAME,
						},
					},
					Model: &libvirtxml.DomainInterfaceModel{
						Type: "virtio",
					},
				},
			},
			Serials: []libvirtxml.DomainSerial{
				{
					Protocol: &libvirtxml.DomainChardevProtocol{
						Type: "pty",
					},
					Target: &libvirtxml.DomainSerialTarget{
						Type: "isa-serial",
						Model: &libvirtxml.DomainSerialTargetModel{
							Name: "isa-serial",
						},
						Port: func() *uint {
							temp := uint(0)
							return &temp
						}(),
					},
				},
				{
					Source: &libvirtxml.DomainChardevSource{
						UNIX: &libvirtxml.DomainChardevSourceUNIX{
							Mode: "bind",
							Path: serialSocket,
						},
					},
					Target: &libvirtxml.DomainSerialTarget{
						Type: "isa-serial",
						Model: &libvirtxml.DomainSerialTargetModel{
							Name: "isa-serial",
						},
						Port: func() *uint {
							temp := uint(1)
							return &temp
						}(),
					},
				},
			},
			Consoles: []libvirtxml.DomainConsole{
				{
					Protocol: &libvirtxml.DomainChardevProtocol{
						Type: "pty",
					},
					Target: &libvirtxml.DomainConsoleTarget{
						Type: "serial",
						Port: func() *uint {
							temp := uint(0)
							return &temp
						}(),
					},
				},
				{
					Source: &libvirtxml.DomainChardevSource{
						UNIX: &libvirtxml.DomainChardevSourceUNIX{
							Mode: "bind",
							Path: consoleSocket,
						},
					},
					Target: &libvirtxml.DomainConsoleTarget{
						Type: "virtio",
						Port: func() *uint {
							temp := uint(0)
							return &temp
						}(),
					},
				},
			},
			Channels: []libvirtxml.DomainChannel{
				{
					Target: &libvirtxml.DomainChannelTarget{
						VirtIO: &libvirtxml.DomainChannelTargetVirtIO{
							Name: "org.qemu.guest_agent.0",
						},
					},
				},
			},
		},
	}
	xmlConfig, err := domConfig.Marshal()
	if err != nil {
		return "", err
	}

	// hacky way to handle <channel>
	// why, libvirtxml, why?
	res := strings.Replace(xmlConfig, "<channel>", `<channel type="unix">`, 1)

	// test the xmlConfig string value
	// fmt.Println(res)

	return res, nil
}

func copyImage(
	instanceName string,
	virtRequest virtualization_model.CreateInstanceRequest,
) error {
	imageMut.Lock()
	defer imageMut.Unlock()

	baseImage := BASE_POOL_DIR + "/" + BASE_IMAGE_NAME
	destinationPath := POOL_DIR + "/" + instanceName + ".qcow2"

	data, err := os.ReadFile(baseImage)
	if err != nil {
		return err
	}

	err = os.WriteFile(destinationPath, data, 0644)
	if err != nil {
		return err
	}

	// resize the qcow2
	resizeCmd := exec.Command("qemu-img", "resize", destinationPath, "+10G")
	// resizeCmd := exec.Command("qemu-img", "resize", destinationPath, fmt.Sprintf("+%dG", virtRequest.Storage))
	resizeCmd.Stderr = os.Stderr
	resizeCmd.Stdout = os.Stdout
	err = resizeCmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func copyEfi(instanceName string) error {
	efiMut.Lock()
	defer efiMut.Unlock()

	destinationPath := NVRAM_DIR + "/" + instanceName + "_VARS.fd"

	data, err := os.ReadFile(NVRAM_TEMPLATE_LOCAL)
	if err != nil {
		return err
	}

	err = os.WriteFile(destinationPath, data, 0644)
	if err != nil {
		return err
	}

	cmd := exec.Command("chown", "qemu:qemu", destinationPath)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func createCloudInitMaster(instanceName string) error {
	cloudInitMut.Lock()
	defer cloudInitMut.Unlock()

	filePath := BASE_POOL_DIR + "/" + "user-data"
	networkPath := BASE_POOL_DIR + "/" + "network-config"
	userDataContent := fmt.Sprintf(`#cloud-config
hostname: %s
locale: en_US.UTF-8
timezone: Asia/Jakarta
users:
- default
- name: user
  groups: sudo
  sudo: ALL=(ALL:ALL) ALL
  plain_text_passwd: user
  lock_passwd: false
  shell: /bin/bash

network:
  version: 2
  ethernets:
    enp1s0:
      addresses:
        - %s/24
      nameservers:
        addresses: [192.168.122.1]
      routes:
        - to: 0.0.0.0/0
          via: 192.168.122.1
          metric: 100

write_files:
- path: /root/service-account.yaml
  content: |
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: admin-user
      namespace: kubernetes-dashboard
- path: /root/role-binding.yaml
  content: |
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: admin-user
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: cluster-admin
    subjects:
    - kind: ServiceAccount
      name: admin-user
      namespace: kubernetes-dashboard

runcmd:
- |
  echo "running command"
  echo "updating and upgrading packages"
        # apt-get update && apt-get upgrade
  
  echo "installing necessary packages"
        # apk add git

  echo "installing k3s"
  curl -sfL https://get.k3s.io | INSTALL_K3S_SKIP_DOWNLOAD=true INSTALL_K3S_EXEC="server --token 12345" sh -s -
        # while [ ! -f /etc/rancher/k3s/k3s.yaml ]; do sleep 1; done

  export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
  echo "export KUBECONFIG=/etc/rancher/k3s/k3s.yaml" >> /etc/profile

  echo "installing helm for kubernetes"
  curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

  echo "creating kubernetes dashboard"
  helm repo add kubernetes-dashboard https://kubernetes.github.io/dashboard/
  helm upgrade --install kubernetes-dashboard kubernetes-dashboard/kubernetes-dashboard --create-namespace --namespace kubernetes-dashboard

  echo "setting up user for kubernetes dashboard"
  k3s kubectl apply -f /root/service-account.yaml -f /root/role-binding.yaml

  echo "writing token and starting the dashboard..."
  echo "waiting until all pods in the kubernetes-dashboard namespaces is running"
  k3s kubectl wait pod --all --for=condition=Ready --namespace=kubernetes-dashboard --timeout=300s
        #rc-service kube-dashboard-port-forward restart

  echo "done"
`, instanceName, MASTER_NODE_IP)

	err := os.WriteFile(filePath, []byte(userDataContent), 0644)
	if err != nil {
		return err
	}

	// create the iso
	cmd := exec.Command("cloud-localds", "-N", networkPath, POOL_DIR+"/"+instanceName+".iso", filePath)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func createCloudInitWorker(instanceName string) error {
	cloudInitMut.Lock()
	defer cloudInitMut.Unlock()

	filePath := BASE_POOL_DIR + "/" + "user-data"
	networkPath := BASE_POOL_DIR + "/" + "network-config"
	userDataContent := fmt.Sprintf(`#cloud-config
hostname: %s
locale: en_US.UTF-8
timezone: Asia/Jakarta
users:
- default
- name: user
  groups: sudo
  sudo: ALL=(ALL:ALL) ALL
  plain_text_passwd: user
  lock_passwd: false
  shell: /bin/bash

network:
  version: 2
  ethernets:
    enp1s0:
      addresses:
        - %s/24
      nameservers:
        addresses: [192.168.122.1]
      routes:
        - to: 0.0.0.0/0
          via: 192.168.122.1
          metric: 100

runcmd:
- |
  echo "running command"
  echo "updating apk and upgrade"
        # apk update && apk upgrade
        # apk add bash
  
  echo "installing k3s"
  curl -sfL https://get.k3s.io | INSTALL_K3S_SKIP_DOWNLOAD=true INSTALL_K3S_EXEC="agent --server https://%s:6443 --token 12345" sh -s -

  echo "done"
        # reboot
`, instanceName, WORKER_NODE_IP, MASTER_NODE_IP)

	err := os.WriteFile(filePath, []byte(userDataContent), 0644)
	if err != nil {
		return err
	}

	// create the iso
	cmd := exec.Command("cloud-localds", "-N", networkPath, POOL_DIR+"/"+instanceName+".iso", filePath)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// NOTE: this function might not needed
// it's needed if we want to configure the network
// whether to be set to static or changing the nameservers, etc.
func createNetworkMaster() error {
	networkMut.Lock()
	defer networkMut.Unlock()

	filePath := BASE_POOL_DIR + "/" + "network-config"
	// NOTE: static address
	userDataContent := fmt.Sprintf(`network:
  version: 2
  ethernets:
    enp1s0:
      addresses:
        - %s/24
      nameservers:
        addresses: [192.168.122.1]
      routes:
        - to: 0.0.0.0/0
          via: 192.168.122.1
          metric: 100
`, MASTER_NODE_IP)

	// NOTE: dynamic address
	//     userDataContent := fmt.Sprintf(`version: 2
	// ethernets:
	//   eth0:
	//     dhcp4: true
	// `)

	err := os.WriteFile(filePath, []byte(userDataContent), 0644)
	if err != nil {
		return err
	}

	return nil
}

// NOTE: this function might not needed
// it's needed if we want to configure the network
// whether to be set to static or changing the nameservers, etc.
func createNetworkWorker() error {
	networkMut.Lock()
	defer networkMut.Unlock()

	filePath := BASE_POOL_DIR + "/" + "network-config"
	// NOTE: static address
	userDataContent := fmt.Sprintf(`network:
  version: 2
  ethernets:
    enp1s0:
      addresses:
        - %s/24
      nameservers:
        addresses: [192.168.122.1]
      routes:
        - to: 0.0.0.0/0
          via: 192.168.122.1
          metric: 100
`, WORKER_NODE_IP)

	// NOTE: dynamic address
	//     userDataContent := fmt.Sprintf(`version: 2
	// ethernets:
	//   eth0:
	//     dhcp4: true
	// `)

	err := os.WriteFile(filePath, []byte(userDataContent), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *LibvirtVirtualization) deleteInstance(
	domainName string,
) {
	domain, err := c.libvirtConnection.LookupDomainByName(domainName)
	if err != nil {
		slog.Error("error getting the domain",
			"error", err,
		)
	}

	// every domain is set to delete when shutdown
	err = domain.Shutdown()
	if err != nil {
		slog.Error("could not shutdown the domain, retrying after this",
			"error", err,
		)

		for i := 1; i <= SHUTDOWN_RETRIES; i++ {
			err = domain.Shutdown()
			if err != nil {
				slog.Error("could not shutdown the domain, retrying after this",
					"error", err,
				)
			}
		}
	}

	// domain files cleanup
	deleteFilesCommand := fmt.Sprintf("rm %s/%s.*", POOL_DIR, domainName)
	cmd := exec.Command("/bin/bash", "-c", deleteFilesCommand)
	err = cmd.Run()
	if err != nil {
		slog.Error("could not clean domain files",
			"error", err,
		)
	}
}
