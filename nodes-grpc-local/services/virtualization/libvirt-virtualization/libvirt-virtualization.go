package libvirt_virtualization

import (
	"fmt"
	"nodes-grpc-local/services/virtualization"
	"os"
	"os/exec"

	"github.com/digitalocean/go-libvirt"
	"github.com/google/uuid"
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

	err := createCloudInit(thisInstanceName)
	if err != nil {
		return err
	}

	err = copyBaseImage(thisInstanceName)
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
	thisInstanceName := uuid.New().String()

	err := createCloudInit(thisInstanceName)
	if err != nil {
		return err
	}

	// domainXmlConfig, err := createBase(thisInstanceName)
	// if err != nil {
	// 	return err
	// }
	//
	// _, err = c.libvirtConnection.DomainCreateXML(domainXmlConfig, libvirt.DomainNone)
	// if err != nil {
	// 	return err
	// }

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

func copyBaseImage(instanceName string) error {
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
	mut.Lock()
	defer mut.Unlock()

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
packages:
- sudo
- findutils
- curl
- util-linux
- dbus
- iproute2
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

// func createCloudInit(instanceName string) error {
// 	metadataContent := fmt.Sprintf(
// 		"instance-id: %s\nlocal-hostname: %s",
// 		instanceName, instanceName,
// 	)
//
// 	return nil
// }
