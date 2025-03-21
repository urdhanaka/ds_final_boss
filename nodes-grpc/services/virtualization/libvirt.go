package virtualization

import (
	"errors"
	"fmt"
	"strings"

	"github.com/digitalocean/go-libvirt"
)

const (
	// cpu threshold
	NODE_BUSY_THRESHOLD float64 = 60

	// libvirt-related error
	LIBVIRT_NOT_INITIALIZED string = "libvirt is not initialized"
	LIBVIRT_POOL_ERROR      string = "storage pool error"

	// storage pool name
	STORAGE_POOL_NAME string = "9b429e57-c9ed-41e5-a87c-ff62de7eb27c"

	// my (urdhanaka) default storage pool name
	MY_STORAGE_POOL_NAME string = "default"
)

type KvmVirtualization struct {
	libvirtConnection *libvirt.Libvirt
}

func NewLibvirtVirt(libvirtConnection *libvirt.Libvirt) *KvmVirtualization {
	return &KvmVirtualization{
		libvirtConnection: libvirtConnection,
	}
}

func (v KvmVirtualization) Spawn() error {
	return nil
}

func (v KvmVirtualization) Stop() error {
	return nil
}

func (v KvmVirtualization) List() error {
	return nil
}

func (v KvmVirtualization) CreateVM() error {
	if v.libvirtConnection == nil {
		return errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	domainDefinition, err := v.libvirtConnection.DomainDefineXML(
		fmt.Sprintf(`
            <domain type="kvm">
                <name>temp</name>
                <memory>786432</memory>
                <currentMemory>786432</currentMemory>
                <vcpu>2</vcpu>
                <os>
                    <type arch="x86_64" machine="pc-q35-9.1">hvm</type>
                    <boot dev="hd"/>
                </os>
                <features>
                    <acpi/>
                    <apic/>
                    <vmport state="off"/>
                </features>
                <cpu mode="host-passthrough"/>
                <clock offset="utc">
                    <timer name="rtc" tickpolicy="catchup"/>
                    <timer name="pit" tickpolicy="delay"/>
                    <timer name="hpet" present="no"/>
                </clock>
                <pm>
                    <suspend-to-mem enabled="no"/>
                    <suspend-to-disk enabled="no"/>
                </pm>
                <devices>
                    <emulator>/usr/bin/qemu-system-x86_64</emulator>
                    <disk type="file" device="cdrom">
                        <driver name="qemu" type="raw"/>
                        <source file="/var/home/urdhanaka/Downloads/alpine-virt-3.21.3-x86_64.iso"/>
                        <target dev="sda" bus="sata"/>
                        <readonly/>
                        <address type="drive" controller="0" bus="0" target="0" unit="0"/>
                    </disk>
                    <disk type="file" device="disk">
                        <driver name="qemu" type="qcow2"/>
                        <source file="/var/lib/libvirt/images/test"/>
                        <target dev="vda" bus="virtio"/>
                        <address type="pci" domain="0x0000" bus="0x04" slot="0x00" function="0x0"/>
                    </disk>
                    <controller type="usb" model="qemu-xhci" ports="15"/>
                    <controller type="pci" model="pcie-root"/>
                    <controller type="pci" model="pcie-root-port"/>
                    <controller type="pci" model="pcie-root-port"/>
                    <controller type="pci" model="pcie-root-port"/>
                    <controller type="pci" model="pcie-root-port"/>
                    <controller type="pci" model="pcie-root-port"/>
                    <controller type="pci" model="pcie-root-port"/>
                    <controller type="pci" model="pcie-root-port"/>
                    <controller type="pci" model="pcie-root-port"/>
                    <controller type="pci" model="pcie-root-port"/>
                    <controller type="pci" model="pcie-root-port"/>
                    <controller type="pci" model="pcie-root-port"/>
                    <controller type="pci" model="pcie-root-port"/>
                    <controller type="pci" model="pcie-root-port"/>
                    <controller type="pci" model="pcie-root-port"/>
                    <interface type="network">
                        <source network="default"/>
                    </interface>
                    <console type="pty"/>
                    <channel type="unix">
                        <source mode="bind"/>
                        <target type="virtio" name="org.qemu.guest_agent.0"/>
                    </channel>
                        <channel type="spicevmc">
                        <target type="virtio" name="com.redhat.spice.0"/>
                    </channel>
                    <input type="tablet" bus="usb"/>
                    <graphics type="spice" port="-1" tlsPort="-1" autoport="yes">
                        <image compression="off"/>
                    </graphics>
                    <sound model="ich9"/>
                    <video>
                        <model type="qxl"/>
                    </video>
                    <redirdev bus="usb" type="spicevmc"/>
                    <redirdev bus="usb" type="spicevmc"/>
                    <memballoon model="virtio"/>
                    <rng model="virtio">
                        <backend model="random">/dev/urandom</backend>
                    </rng>
                </devices>
            </domain>
            `))
	if err != nil {
		return err
	}

	err = v.libvirtConnection.DomainCreate(domainDefinition)
	if err != nil {
		return err
	}

	return nil
}

func (v KvmVirtualization) CreateVolume() error {
	if v.libvirtConnection == nil {
		return errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	storagePool, err := v.GetStoragePool("default")
	if err != nil {
		return err
	}

	storageVolCreateFlags := libvirt.StorageVolCreateFlags(0)
	_, err = v.libvirtConnection.StorageVolCreateXML(storagePool,
		`
        <volume>
            <name>test</name>
            <capacity>21474836480</capacity>
            <allocation>0</allocation>
            <target>
                <format type='qcow2'/>
            </target>
        </volume>
        `, storageVolCreateFlags)
	if err != nil {
		return err
	}

	return nil
}

func (v KvmVirtualization) CreatePool() error {
	if v.libvirtConnection == nil {
		return errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	storagePool, err := v.GetStoragePool("")
	if err != nil {
		// check if storage pool is not available
		if strings.Contains(err.Error(), "no storage pool with matching name") {
			_, err = v.libvirtConnection.StoragePoolCreateXML(
				fmt.Sprintf(`
                    <pool type="dir">
                    <name>%s</name>
                    <target>
                    <path>/var/lib/libvirt/images</path>
                    </target>
                    </pool>
                    `, MY_STORAGE_POOL_NAME), libvirt.StoragePoolCreateNormal)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	isActive, err := v.libvirtConnection.StoragePoolIsActive(storagePool)
	if err != nil {
		return err
	}

	if isActive == 0 { // not running
		err := v.libvirtConnection.StoragePoolCreate(storagePool, libvirt.StoragePoolCreateNormal)
		if err != nil {
			return err
		}
	}

	return nil
}

func (v KvmVirtualization) ListStoragePools() ([]libvirt.StoragePool, error) {
	if v.libvirtConnection == nil {
		return nil, errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	// for now, only list the storage pools
	// that already defined and either online or offline
	listStoragePoolsFlags := libvirt.ConnectListStoragePoolsActive
	listStoragePoolsFlags |= libvirt.ConnectListStoragePoolsInactive
	storagePools, _, err := v.libvirtConnection.ConnectListAllStoragePools(
		1,
		listStoragePoolsFlags,
	)
	if err != nil {
		return nil, err
	}

	return storagePools, nil
}

func (v KvmVirtualization) GetStoragePool(name string) (libvirt.StoragePool, error) {
	var storagePool libvirt.StoragePool
	var err error

	if v.libvirtConnection == nil {
		return libvirt.StoragePool{}, errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	if name == "" {
		storagePool, err = v.libvirtConnection.StoragePoolLookupByName(MY_STORAGE_POOL_NAME)
	} else {
		storagePool, err = v.libvirtConnection.StoragePoolLookupByName(name)
	}
	if err != nil {
		return libvirt.StoragePool{}, err
	}

	return storagePool, nil
}
