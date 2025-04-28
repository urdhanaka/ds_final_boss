package libvirt_virtualization

import (
	"fmt"
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

func (c *LibvirtVirtualization) CreateMaster() error {
	thisInstanceName := generateRandom(10)

	err := createNetworkMaster()
	if err != nil {
		return err
	}

	err = createCloudInitMaster(thisInstanceName)
	if err != nil {
		return err
	}

	err = copyImage(thisInstanceName)
	if err != nil {
		return err
	}

	domainXmlConfig, err := createBase(thisInstanceName)
	if err != nil {
		return err
	}

	_, err = c.libvirtConnection.DomainCreateXML(domainXmlConfig, libvirt.DomainNone)
	if err != nil {
		return err
	}

	return nil
}

func (c *LibvirtVirtualization) CreateWorker() error {
	thisInstanceName := generateRandom(10)

	err := createNetworkWorker()
	if err != nil {
		return err
	}

	err = createCloudInitWorker(thisInstanceName)
	if err != nil {
		return err
	}

	err = copyImage(thisInstanceName)
	if err != nil {
		return err
	}

	domainXmlConfig, err := createBase(thisInstanceName)
	if err != nil {
		return err
	}

	_, err = c.libvirtConnection.DomainCreateXML(domainXmlConfig, libvirt.DomainNone)
	if err != nil {
		return err
	}

	return nil
}

func createBase(instanceName string) (string, error) {
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
			Value: 2097152, // WARN: hardcoded
		},
		VCPU: &libvirtxml.DomainVCPU{
			Value: 4, // WARN: hardcoded
		},
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{
				Arch:    "x86_64",
				Machine: "q35",
				Type:    "hvm",
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
		},
	}
	xmlConfig, err := domConfig.Marshal()
	if err != nil {
		return "", err
	}

	return xmlConfig, nil
}

func copyImage(instanceName string) error {
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
	resizeCmd.Stderr = os.Stderr
	resizeCmd.Stdout = os.Stdout
	err = resizeCmd.Run()
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
- doas: [permit nopass user]
  name: user
  plain_text_passwd: user
  lock_passwd: false
  shell: /bin/sh

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
  shell: /bin/sh

runcmd:
- |
  echo "running command"
        # echo "configuring cgroup..."
        # touch /boot/cmdline.txt
        # echo "cgroup_memory=1 cgroup_enable=memory" >> /boot/cmdline.txt
  
        # echo "configuring /etc/resolv.conf"
  # echo "nameserver 192.168.122.1" >> /etc/resolv.conf

  echo "updating apk and upgrade"
  apk update && apk upgrade
  
  # echo "installing necessary packages"
  # apk add sudo findutils iptables curl util-linux dbus iproute2 bash openssl git

  echo "installing k3s"
  curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="server --token 12345" sh -s -

  echo "installing helm for kubernetes"
  echo "export KUBECONFIG=/etc/rancher/k3s/k3s.yaml" >> /etc/profile
  source /etc/profile
  curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

  echo "creating kubernetes dashboard"
  export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
  helm repo add kubernetes-dashboard https://kubernetes.github.io/dashboard/
  helm upgrade --install kubernetes-dashboard kubernetes-dashboard/kubernetes-dashboard --create-namespace --namespace kubernetes-dashboard

  echo "done"
        # reboot
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
  plain_text_passwd: user
  lock_passwd: false
  shell: /bin/sh

runcmd:
- |
  echo "running command"
        #echo "configuring cgroup..."
        #touch /boot/cmdline.txt
        #echo "cgroup_memory=1 cgroup_enable=memory" >> /boot/cmdline.txt
  
        #echo "configuring /etc/resolv.conf"
        #echo "nameserver 192.168.122.1" >> /etc/resolv.conf

  echo "updating apk and upgrade"
  apk update && apk upgrade
  
        #echo "installing necessary packages"
        #apk add sudo findutils iptables curl util-linux dbus iproute2 bash git

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

// func createCloudInit(instanceName string) error {
// 	metadataContent := fmt.Sprintf(
// 		"instance-id: %s\nlocal-hostname: %s",
// 		instanceName, instanceName,
// 	)
//
// 	return nil
// }
