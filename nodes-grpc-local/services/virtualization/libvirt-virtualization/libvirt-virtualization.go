package libvirt_virtualization

import (
	"context"
	"fmt"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"
	"nodes-grpc-local/services/virtualization"
	"os"
	"os/exec"

	"github.com/digitalocean/go-libvirt"
	"libvirt.org/go/libvirtxml"
)

type LibvirtVirtualization struct {
	libvirtConnection *libvirt.Libvirt
}

func NewLibvirtVirtualization(
	libvirtConnection *libvirt.Libvirt,
) virtualization.VirtualizationInterface {
	return &LibvirtVirtualization{
		libvirtConnection: libvirtConnection,
	}
}

func (c *LibvirtVirtualization) CreateMaster(
	ctx context.Context,
	virtRequest virtualization_model.CreateInstanceRequest,
) error {
	thisInstanceName := generateRandom(10)

	err := createNetworkMaster()
	if err != nil {
		return err
	}

	err = createCloudInitMaster(thisInstanceName)
	if err != nil {
		return err
	}

	err = copyImage(thisInstanceName, virtRequest)
	if err != nil {
		return err
	}

    err = copyEfi(thisInstanceName)
    if err != nil {
        return err
    }

	domainXmlConfig, err := createBase(thisInstanceName, virtRequest)
	if err != nil {
		return err
	}

	_, err = c.libvirtConnection.DomainCreateXML(domainXmlConfig, libvirt.DomainNone)
	if err != nil {
		return err
	}

	return nil
}

func (c *LibvirtVirtualization) CreateWorker(
	ctx context.Context,
	virtRequest virtualization_model.CreateInstanceRequest,
) error {
	thisInstanceName := generateRandom(10)

	err := createNetworkWorker()
	if err != nil {
		return err
	}

	err = createCloudInitWorker(thisInstanceName)
	if err != nil {
		return err
	}

	err = copyImage(thisInstanceName, virtRequest)
	if err != nil {
		return err
	}

    err = copyEfi(thisInstanceName)
    if err != nil {
        return err
    }

	domainXmlConfig, err := createBase(thisInstanceName, virtRequest)
	if err != nil {
		return err
	}

	_, err = c.libvirtConnection.DomainCreateXML(domainXmlConfig, libvirt.DomainNone)
	if err != nil {
		return err
	}

	return nil
}

func createBase(instanceName string, instanceConfig virtualization_model.CreateInstanceRequest) (string, error) {
	instanceStorage := POOL_DIR + "/" + instanceName + ".qcow2"
	seedFile := POOL_DIR + "/" + instanceName + ".iso"

	domConfig := &libvirtxml.Domain{
		Type: "kvm",
		Name: instanceName,
		Metadata: &libvirtxml.DomainMetadata{
			XML: `<libosinfo:libosinfo xmlns:libosinfo="http://libosinfo.org/xmlns/libvirt/domain/1.0">
                    <libosinfo:os id="http://alpinelinux.org/alpinelinux/3.21"/>
                  </libosinfo:libosinfo>`,
		},
		Memory: &libvirtxml.DomainMemory{
			Value: 2097152, // WARN: hadrcoded
			// Unit:  "GB",
		},
		VCPU: &libvirtxml.DomainVCPU{
			Value: 4, // WARN: hardcoded
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
				Path:     "/usr/share/edk2/ovmf/OVMF_CODE.fd",
			},
			NVRam: &libvirtxml.DomainNVRam{
				NVRam:          fmt.Sprintf("/var/lib/libvirt/qemu/nvram/_VARS.fd", instanceName), // TODO:
				Template:       "/usr/share/edk2/ovmf/OVMF_VARS.fd",
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
							Network: "default",
							Bridge:  "virbr0",
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
					Protocol: &libvirtxml.DomainChardevProtocol{
						Type: "unix",
					},
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

	return xmlConfig, nil
}

func copyImage(instanceName string, virtRequest virtualization_model.CreateInstanceRequest) error {
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
	// resizeCmd := exec.Command("qemu-img", "resize", destinationPath, fmt.Sprintf("+%dG", virtRequest.Storage))
	resizeCmd := exec.Command("qemu-img", "resize", destinationPath, "+10G")
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

	data, err := os.ReadFile(NVRAM_TEMPLATE)
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

func createCloudInit(instanceName string) error {
	cloudInitMut.Lock()
	defer cloudInitMut.Unlock()

	filePath := BASE_POOL_DIR + "/" + "user-data"
	userDataContent := fmt.Sprintf(`#cloud-config
hostname: %s
users:
- default

package_update: true
package_upgrade: true
packages:
- sudo
- findutils
- iptables
- curl
- util-linux
- dbus
- iproute2

runcmd:
- |
  echo "running command"
  echo "configuring cgroup..."
  touch /boot/cmdline.txt
  echo "cgroup_memory=1 cgroup_enable=memory" >> /boot/cmdline.txt

  echo "installing k3s"
  curl -sfL https://get.k3s.io | sh -

  echo "done"
  reboot
`, instanceName)

	err := os.WriteFile(filePath, []byte(userDataContent), 0644)
	if err != nil {
		return err
	}

	// create the iso
	cmd := exec.Command("cloud-localds", POOL_DIR+"/"+instanceName+".iso", filePath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
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
locale: en_US
timezone: Asia/Jakarta
users:
- default
- doas: [permit nopass user]
  name: user
  groups: wheel
  plain_text_passwd: user
  lock_passwd: false
  shell: /bin/bash

write_files:
- path: /etc/init.d/kube-dashboard-proxy
  permissions: '0755'
  owner: root:root
  content: |
  #!/sbin/openrc-run
  description="Kubernetes dashboard proxy"

  command="/usr/local/bin/k3s"
  command_args="kubectl proxy --port=8001 --address=0.0.0.0 --accept-hosts=\(.*?\))
  pidfile="/run/kubectl-proxy.pid"
  command_background="yes"

runcmd:
- |
  echo "running command"
  echo "updating apk and upgrade"
  apk update && apk upgrade
  
  # echo "installing necessary packages"
  # apk add sudo findutils iptables curl util-linux dbus iproute2 bash openssl git mount

  echo "installing k3s"
  curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="server --token 12345" sh -s -

  while [ ! -f /etc/rancher/k3s/k3s.yaml ]; do sleep 1; done
  mkdir /home/user/.kube
  cp /etc/rancher/k3s/k3s.yaml /home/user/.kube/config
  chown user:user /home/user/.kube/config

  export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
  echo "export KUBECONFIG=/etc/rancher/k3s/k3s.yaml" >> /etc/profile

  echo "installing helm for kubernetes"
  curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

  echo "creating kubernetes dashboard"
  helm repo add kubernetes-dashboard https://kubernetes.github.io/dashboard/
  helm upgrade --install kubernetes-dashboard kubernetes-dashboard/kubernetes-dashboard --create-namespace --namespace kubernetes-dashboard

  echo "setting up user for kubernetes dashboard"
  cat <<EOF | k3s kubectl apply -f -
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: admin-user
      namespace: kubernetes-dashboard
    ---
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
  EOF

  echo "writing token and starting the dashboard..."
  rc-service kube-dashboard-proxy restart
  kubectl -n kubernetes-dashboard create token admin-user >> /home/user/.token

  echo "done"
`, instanceName)

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

func createCloudInitWorker(instanceName string) error {
	cloudInitMut.Lock()
	defer cloudInitMut.Unlock()

	filePath := BASE_POOL_DIR + "/" + "user-data"
	networkPath := BASE_POOL_DIR + "/" + "network-config"
	userDataContent := fmt.Sprintf(`#cloud-config
hostname: %s
locale: en_US
timezone: Asia/Jakarta
users:
- default
- doas: [permit nopass user]
  name: user
  groups: wheel
  plain_text_passwd: user
  lock_passwd: false
  shell: /bin/bash

runcmd:
- |
  echo "running command"
  echo "updating apk and upgrade"
  apk update && apk upgrade
  
  echo "installing k3s"
  curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="agent --server https://192.168.122.49:6443 --token 12345" sh -s -

  echo "done"
  reboot
`, instanceName)

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

func createNetworkMaster() error {
	networkMut.Lock()
	defer networkMut.Unlock()

	filePath := BASE_POOL_DIR + "/" + "network-config"
	// NOTE: static address
	userDataContent := `version: 2
ethernets:
  eth0:
    addresses:
      - 192.168.122.49/24
    nameservers:
      addresses: [192.168.122.1]
    routes:
      - to: 0.0.0.0/0
        via: 192.168.122.1
        metric: 100
`

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

func createNetworkWorker() error {
	networkMut.Lock()
	defer networkMut.Unlock()

	filePath := BASE_POOL_DIR + "/" + "network-config"
	// NOTE: static address
	userDataContent := fmt.Sprintf(`version: 2
ethernets:
  eth0:
    addresses:
      - 192.168.122.50/24
    nameservers:
      search: [if.its.ac.id]
      addresses: [202.46.129.3]
    routes:
      - to: 0.0.0.0/0
        via: 192.168.122.1
        metric: 100
`)

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
