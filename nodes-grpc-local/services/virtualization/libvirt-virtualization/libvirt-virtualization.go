package libvirt_virtualization

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"
	"os"
	"os/exec"
	"strings"
	"time"

	"libvirt.org/go/libvirt"
	"libvirt.org/go/libvirtxml"
)

// these hardcoded IP are for testing only
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
	libvirtConnection *libvirt.Connect
}

func NewLibvirtVirtualization(
	libvirtConnection *libvirt.Connect,
) *LibvirtVirtualization {
	return &LibvirtVirtualization{
		libvirtConnection: libvirtConnection,
	}
}

// wrapper for create master and create worker
// this function will be used in queue
func (c *LibvirtVirtualization) CreateInstance(
	ctx context.Context,
	virtRequest *virtualization_model.CreateInstanceRequest,
) (*virtualization_model.VirtCreateInstanceResponse, error) {
	if virtRequest.IsMaster {
		return c.createMaster(ctx, virtRequest)
	} else {
		return c.createWorker(ctx, virtRequest)
	}
}

func (c *LibvirtVirtualization) createMaster(
	ctx context.Context,
	virtRequest *virtualization_model.CreateInstanceRequest,
) (*virtualization_model.VirtCreateInstanceResponse, error) {
	createRes := new(virtualization_model.VirtCreateInstanceResponse)
	createRes.Status = false

	thisInstanceName := virtRequest.Name

	slog.Info(fmt.Sprintf("master node name is %s", thisInstanceName))
	slogFunction(virtRequest.Name, thisInstanceName, "creating master instance", nil)

	// create network
	slogFunction(virtRequest.Name, thisInstanceName, "creating node network", nil)
	err := createNetwork()
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "could not create node network", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}

	// create cloudinit configuration
	slogFunction(virtRequest.Name, thisInstanceName, "creating node cloud-init configuration", nil)
	err = createCloudInitMaster(thisInstanceName)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "could not create node cloud-init configuration", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}

	// creating the image
	slogFunction(virtRequest.Name, thisInstanceName, "creating base image for instance", nil)
	err = copyImage(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "could not create node image", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}

	// configure the EFI
	slogFunction(virtRequest.Name, thisInstanceName, "creating instance EFI file", nil)
	err = copyEfi(thisInstanceName)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "could not create node EFI", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}

	// base xml for the vm
	slogFunction(virtRequest.Name, thisInstanceName, "creating instance base xml", nil)
	domainXmlConfig, err := createBase(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "could not create node base xml", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}

	// spawning
	slogFunction(virtRequest.Name, thisInstanceName, "spawning node", nil)
	dom, err := c.libvirtConnection.DomainCreateXML(domainXmlConfig, libvirt.DomainCreateFlags(0))
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "could not spawn node", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}

	slogFunction(virtRequest.Name, thisInstanceName, "waiting until the vm is ready..", nil)
	time.Sleep(15 * time.Second)
	waitCloudInitCmd := "cloud-init status --wait"
	_, err = guestAgentExecStatus(dom, waitCloudInitCmd)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "error waiting cloud-init process", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}

	// creating token
	kubeCreateTokenCmd := "k3s kubectl -n kubernetes-dashboard create token admin-user"
	createTokenStatus, err := guestAgentExecStatus(dom, kubeCreateTokenCmd)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "error creating kubernetes dashboard token", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}
	// handle base 64 of the guest agent result
	decodedTokenBytes, err := base64.StdEncoding.DecodeString(createTokenStatus.Return.OutData)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "error decoding kubernetes dashboard bytes", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}

	// getting the IP address
	// DO NOT TOUCH THE sed SEQUENCE
	ipAddressCmd := `ip -f inet addr show enp1s0 | sed -En -e 's/.*inet ([0-9.]+).*/\\1/p'`
	ipAddressStatus, err := guestAgentExecStatus(dom, ipAddressCmd)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "error getting master instance ip address", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}
	// handle base 64 of the guest agent result
	decodedIpAddressBytes, err := base64.StdEncoding.DecodeString(ipAddressStatus.Return.OutData)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "error decoding kubernetes dashboard bytes", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}

	createRes.Status = true
	createRes.DashboardToken = string(decodedTokenBytes)
	createRes.MasterIpAddress = string(decodedIpAddressBytes)

	fmt.Println("create master response: ", createRes)

	return createRes, nil
}

func (c *LibvirtVirtualization) createWorker(
	ctx context.Context,
	virtRequest *virtualization_model.CreateInstanceRequest,
) (*virtualization_model.VirtCreateInstanceResponse, error) {
	createRes := new(virtualization_model.VirtCreateInstanceResponse)
	createRes.Status = false

	thisInstanceName := virtRequest.Name

	slog.Info(fmt.Sprintf("worker node name is %s", thisInstanceName))
	slogFunction(virtRequest.Name, thisInstanceName, "creating worker instance", nil)

	// create network
	slogFunction(virtRequest.Name, thisInstanceName, "creating instance network", nil)
	err := createNetwork()
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "could not create node network", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}

	// create cloudinit configuration
	slogFunction(virtRequest.Name, thisInstanceName, "creating node cloud-init configuration", nil)
	err = createCloudInitWorker(thisInstanceName)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "could not create node cloud-init configuration", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}

	// creating the image
	slogFunction(virtRequest.Name, thisInstanceName, "creating node image", nil)
	err = copyImage(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "could not create node image", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}

	// configure the EFI
	slogFunction(virtRequest.Name, thisInstanceName, "creating node EFI", nil)
	err = copyEfi(thisInstanceName)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "could not create node EFI", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}

	// base xml for the vm
	slogFunction(virtRequest.Name, thisInstanceName, "creating node base xml", nil)
	domainXmlConfig, err := createBase(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "could not create node base xml", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}

	// spawning
	slogFunction(virtRequest.Name, thisInstanceName, "spawning node", nil)
	dom, err := c.libvirtConnection.DomainCreateXML(domainXmlConfig, libvirt.DOMAIN_NONE)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "could not spawn node", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}
	slogFunction(virtRequest.Name, thisInstanceName, "waiting until the vm is ready..", nil)
	time.Sleep(15 * time.Second)
	waitCloudInitCmd := "cloud-init status --wait"
	_, err = guestAgentExecStatus(dom, waitCloudInitCmd)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "could not spawn node", err)
		c.deleteInstance(thisInstanceName)

		return createRes, err
	}

	createRes.Status = true

	return createRes, nil
}

func createBase(
	instanceName string,
	instanceConfig *virtualization_model.CreateInstanceRequest,
) (string, error) {
	instanceStorage := POOL_DIR + "/" + instanceName + ".qcow2"
	seedFile := POOL_DIR + "/" + instanceName + ".iso"
	logSocket := INSTANCE_LOGS_DIR + "/" + instanceName + ".sock"

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
				// Template: NVRAM_TEMPLATE_LOCAL, // NOTE: only works in local
				Template:       NVRAM_TEMPLATE, // NOTE: directory according to ubuntu 24.04
				TemplateFormat: "raw",
				Format:         "raw",
			},
			BootDevices: []libvirtxml.DomainBootDevice{
				{
					Dev: "hd",
				},
			},
			BIOS: &libvirtxml.DomainBIOS{
				UseSerial: "yes",
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
		OnCrash:    "restart",
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
						Port: func() *uint {
							temp := uint(0)
							return &temp
						}(),
						Model: &libvirtxml.DomainSerialTargetModel{
							Name: "isa-serial",
						},
					},
				},
				{
					Source: &libvirtxml.DomainChardevSource{
						UNIX: &libvirtxml.DomainChardevSourceUNIX{
							Mode: "bind",
							Path: logSocket,
						},
					},
					Target: &libvirtxml.DomainSerialTarget{
						Type: "isa-serial",
						Port: func() *uint {
							temp := uint(1)
							return &temp
						}(),
						Model: &libvirtxml.DomainSerialTargetModel{
							Name: "isa-serial",
						},
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
							temp := uint(1)
							return &temp
						}(),
					},
				},
				// {
				// 	Source: &libvirtxml.DomainChardevSource{
				// 		UNIX: &libvirtxml.DomainChardevSourceUNIX{
				// 			Mode: "bind",
				// 			Path: logSocket,
				// 		},
				// 	},
				// 	Target: &libvirtxml.DomainConsoleTarget{
				// 		Type: "serial",
				// 		Port: func() *uint {
				// 			temp := uint(1)
				// 			return &temp
				// 		}(),
				// 	},
				// },
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
	virtRequest *virtualization_model.CreateInstanceRequest,
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
	// resizeCmd := exec.Command("qemu-img", "resize", destinationPath, "+10G")
	resizeCmd := exec.Command("qemu-img", "resize", destinationPath, fmt.Sprintf("+%dG", virtRequest.Storage))
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
- path: /etc/systemd/system/kube-dashboard.service
  content: |
    [Unit]
    Description=Kubernetes dashboard
    Wants=network-online.target
    After=k3s.service

    [Install]
    WantedBy=multi-user.target

    [Service]
    Type=simple
    User=root
    Restart=always
    RestartSec=5s
    ExecStart=/usr/local/bin/k3s \
        kubectl -n kubernetes-dashboard \
        port-forward svc/kubernetes-dashboard-kong-proxy \
        8443:443 --address 0.0.0.0 \

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

        #echo "installing helm for kubernetes"
        #curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

  echo "creating kubernetes dashboard"
  helm repo add kubernetes-dashboard https://kubernetes.github.io/dashboard/
  helm upgrade --install kubernetes-dashboard kubernetes-dashboard/kubernetes-dashboard --create-namespace --namespace kubernetes-dashboard

  echo "setting up user for kubernetes dashboard"
  k3s kubectl apply -f /root/service-account.yaml -f /root/role-binding.yaml

  echo "writing token and starting the dashboard..."
  echo "waiting until all pods in the kubernetes-dashboard namespaces is running"
  k3s kubectl wait pod --all --for=condition=Ready --namespace=kubernetes-dashboard --timeout=-1s
  systemctl start kube-dashboard.service

  echo "done"
`, instanceName)

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

func createNetwork() error {
    networkMut.Lock()
    defer networkMut.Unlock()

    filePath := BASE_POOL_DIR + "/" + "network-config"
    userDataContent := `network:
  version: 2
  ethernets:
    enp1s0:
      dhcp4: true`

    err := os.WriteFile(filePath, []byte(userDataContent), 0644)
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
	// 	userDataContent := fmt.Sprintf(`network:
	//   version: 2
	//   ethernets:
	//     enp1s0:
	//       addresses:
	//         - %s/24
	//       nameservers:
	//         addresses: [192.168.122.1]
	//       routes:
	//         - to: 0.0.0.0/0
	//           via: 192.168.122.1
	//           metric: 100
	// `, MASTER_NODE_IP)

	// NOTE: dynamic address
	userDataContent := fmt.Sprintf(`network:
  version: 2
  ethernets:
    enp1s0:
	  dhcp4: true`)

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
// 	userDataContent := fmt.Sprintf(`network:
//   version: 2
//   ethernets:
//     enp1s0:
//       addresses:
//         - %s/24
//       nameservers:
//         addresses: [192.168.122.1]
//       routes:
//         - to: 0.0.0.0/0
//           via: 192.168.122.1
//           metric: 100
// `, WORKER_NODE_IP)

	// NOTE: dynamic address
	userDataContent := fmt.Sprintf(`network:
  version: 2
  ethernets:
    enp1s0:
	  dhcp4: true`)

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
	deleteFilesCommand = fmt.Sprintf("rm %s/%s.*", NVRAM_DIR, domainName)
	cmd = exec.Command("/bin/bash", "-c", deleteFilesCommand)
	err = cmd.Run()
	if err != nil {
		slog.Error("could not clean domain files",
			"error", err,
		)
	}
}

func guestAgentExecStatus(
	dom *libvirt.Domain,
	execString string,
) (*ExecStatusQemuGuestAgent, error) {
	res := new(ExecStatusQemuGuestAgent)

	execCmd := fmt.Sprintf(`{"execute":"guest-exec","arguments":{"path":"/bin/bash","arg":["-c", "%s"],"capture-output":true}}`,
		execString)
	cmdPid, err := dom.QemuAgentCommand(execCmd, libvirt.DOMAIN_QEMU_AGENT_COMMAND_BLOCK, 0)
	if err != nil {
		return res, err
	}

	// handle the pid
	pidStruct := new(PidQemuGuestAgent)
	err = json.Unmarshal([]byte(cmdPid), pidStruct)
	if err != nil {
		return res, err
	}

	for {
		cmd := fmt.Sprintf(`{"execute":"guest-exec-status","arguments":{"pid":%d}}`, pidStruct.Return.PID)
		status, err := dom.QemuAgentCommand(cmd, libvirt.DOMAIN_QEMU_AGENT_COMMAND_BLOCK, 0)
		if err != nil {
			return res, err
		}

		fmt.Println(status)

		err = json.Unmarshal([]byte(status), res)
		if err != nil {
			return res, err
		}

		fmt.Println(res)

		if res.Return.Exited {
			break
		}

		time.Sleep(10 * time.Second)
	}

	return res, nil
}
