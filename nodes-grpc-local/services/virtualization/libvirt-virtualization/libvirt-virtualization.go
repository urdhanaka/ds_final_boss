package libvirt_virtualization

import (
	"context"
	"fmt"
	"log/slog"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"
	"nodes-grpc-local/services/virtualization"
	"os"
	"os/exec"
	"strings"
	"time"

	"libvirt.org/go/libvirt"
	"libvirt.org/go/libvirtxml"
)

const (
	MASTER_NODE_IP     = "192.168.122.49"
	WORKER_NODE_IP     = "192.168.122.50"
	CLOUD_INIT_TIMEOUT = 60
)

type LibvirtVirtualization struct {
	libvirtConnection *libvirt.Connect
}

func NewLibvirtVirtualization(
	libvirtConnection *libvirt.Connect,
) virtualization.VirtualizationInterface {
	return &LibvirtVirtualization{
		libvirtConnection: libvirtConnection,
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

	slog.Info(fmt.Sprintf("master node name is %s", thisInstanceName))
	slogFunction(thisInstanceName, "creating master instance", nil)

	// create network
	slogFunction(thisInstanceName, "creating node network", nil)
	err := createNetworkMaster()
	if err != nil {
		slogFunction(thisInstanceName, "could not create node network", err)

		return err
	}

	// create cloudinit configuration
	slogFunction(thisInstanceName, "creating node cloud-init configuration", nil)
	err = createCloudInitMaster(thisInstanceName)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node cloud-init configuration", err)

		return err
	}

	// creating the image
	slogFunction(thisInstanceName, "creating node image", nil)
	err = copyImage(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node image", err)

		return err
	}

	// configure the EFI
	slogFunction(thisInstanceName, "creating node EFI", nil)
	err = copyEfi(thisInstanceName)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node EFI", err)

		return err
	}

	// base xml for the vm
	slogFunction(thisInstanceName, "creating node base xml", nil)
	domainXmlConfig, err := createBase(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node base xml", err)

		return err
	}

	// spawning
	slogFunction(thisInstanceName, "spawning node", nil)
	dom, err := c.libvirtConnection.DomainCreateXML(domainXmlConfig, libvirt.DomainCreateFlags(0))
	if err != nil {
		slogFunction(thisInstanceName, "could not spawn node", err)

		return err
	}
	slogFunction(thisInstanceName, "waiting until the vm is ready..", nil)
    time.Sleep(15 * time.Second)
	waitCloudInitCmd := `{"execute":"guest-exec","arguments":{"path":"/bin/bash","arg":["-c", "cloud-init status --wait"],"capture-output":true}}`
	result, err := dom.QemuAgentCommand(waitCloudInitCmd, libvirt.DOMAIN_QEMU_AGENT_COMMAND_BLOCK, 0)
	if err != nil {
		slogFunction(thisInstanceName, "could not spawn node", err)

		return err
	}

	fmt.Println(result)

	return nil
}

func (c *LibvirtVirtualization) createWorker(
	ctx context.Context,
	virtRequest virtualization_model.CreateInstanceRequest,
) error {
	thisInstanceName := virtRequest.Name

	slog.Info(fmt.Sprintf("worker node name is %s", thisInstanceName))
	slogFunction(thisInstanceName, "creating worker instance", nil)

	slogFunction(thisInstanceName, "creating node network", nil)
	err := createNetworkWorker()
	if err != nil {
		slogFunction(thisInstanceName, "could not create node network", err)

		return err
	}

	slogFunction(thisInstanceName, "creating node cloud-init configuration", nil)
	err = createCloudInitWorker(thisInstanceName)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node cloud-init configuration", err)

		return err
	}

	slogFunction(thisInstanceName, "creating node image", nil)
	err = copyImage(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node image", err)

		return err
	}

	slogFunction(thisInstanceName, "creating node EFI", nil)
	err = copyEfi(thisInstanceName)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node EFI", err)

		return err
	}

	slogFunction(thisInstanceName, "creating node base xml", nil)
	domainXmlConfig, err := createBase(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(thisInstanceName, "could not create node base xml", err)

		return err
	}

	slogFunction(thisInstanceName, "spawning node", nil)
	_, err = c.libvirtConnection.DomainCreateXML(domainXmlConfig, libvirt.DOMAIN_NONE)
	if err != nil {
		slogFunction(thisInstanceName, "could not spawn node", err)

		return err
	}

	return nil
}

func (c *LibvirtVirtualization) StopInstance(
	ctx context.Context,
	instance virtualization_model.Instance,
) error {
	slogFunction(instance.Name, "shutting down instance...", nil)

	dom, err := c.libvirtConnection.LookupDomainByName(instance.Name)
	if err != nil {
		slogFunction(instance.Name, "could not shut down domain", err)

		return err
	}

	err = dom.Shutdown()
	if err != nil {
		slogFunction(instance.Name, "could not shut down domain", err)

		return err
	}

	return nil
}

func createBase(
	instanceName string,
	instanceConfig virtualization_model.CreateInstanceRequest,
) (string, error) {
	instanceStorage := POOL_DIR + "/" + instanceName + ".qcow2"
	seedFile := POOL_DIR + "/" + instanceName + ".iso"

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
				Path:     LOADER_LOCAL, // NOTE
				// Path:     "/usr/share/OVMF/OVMF_CODE_4M.fd",
			},
			NVRam: &libvirtxml.DomainNVRam{
				NVRam:    fmt.Sprintf("/var/lib/libvirt/qemu/nvram/%s_VARS.fd", instanceName),
				Template: NVRAM_TEMPLATE_LOCAL, // NOTE: only works in local
				// Template:       NVRAM_TEMPLATE_LOCAL, // NOTE: directory according to ubuntu 24.04
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
			Consoles: []libvirtxml.DomainConsole{
				{
					TTY: "pty",
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
  curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="server --token 12345" sh -s -
  while [ ! -f /etc/rancher/k3s/k3s.yaml ]; do sleep 1; done

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
        #k3s kubectl wait pod --all --for=condition=Ready --namespace=kubernetes-dashboard --timeout=120s
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
  curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="agent --server https://%s:6443 --token 12345" sh -s -

  echo "done"
        # reboot
`, instanceName, WORKER_NODE_IP, MASTER_NODE_IP)

	err := os.WriteFile(filePath, []byte(userDataContent), 0644)
	if err != nil {
		return err
	}

	// create the iso
	cmd := exec.Command("cloud-localds", "-N", networkPath, POOL_DIR+"/"+instanceName+".iso", filePath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
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
