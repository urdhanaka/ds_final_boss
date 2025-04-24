package libvirt_virtualization

import (
	"fmt"
	"nodes-grpc-local/services/virtualization"

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
	return nil
}

func (c *LibvirtVirtualization) CreateWorker() error {
	return nil
}

func (c *LibvirtVirtualization) createBase() error {
	domConfig := &libvirtxml.Domain{
		Type: "kvm",
		OS: &libvirtxml.DomainOS{
			Type: &libvirtxml.DomainOSType{
				Arch: "x86_64",
			},
		},
		Features: &libvirtxml.DomainFeatureList{
			ACPI: &libvirtxml.DomainFeature{},
			APIC: &libvirtxml.DomainFeatureAPIC{},
		},
		OnPoweroff: "destroy",
		OnCrash:    "destroy",
	}
	xmlConfig, err := domConfig.Marshal()
	if err != nil {
		return err
	}

	fmt.Println(xmlConfig)

	return nil
}
