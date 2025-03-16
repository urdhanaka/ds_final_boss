package virtualization

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/digitalocean/go-libvirt"
)

const (
	// cpu threshold
	NODE_BUSY_THRESHOLD float64 = 60

	// libvirt-related error
	LIBVIRT_NOT_INITIALIZED string = "libvirt is not initialized"
	LIBVIRT_POOL_ERROR      string = "storage pool error"
)

// initialize libvirt connection using qemu
func initKvmConnection() *libvirt.Libvirt {
	uri, _ := url.Parse(string(libvirt.QEMUSystem))
	connection, err := libvirt.ConnectToURI(uri)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to qemu: %s", err.Error())
	}

	return connection
}

type KvmVirtualization struct {
	connection *libvirt.Libvirt
}

func NewLibvirtVirt() *KvmVirtualization {
	return &KvmVirtualization{
		connection: initKvmConnection(),
	}
}

func (v KvmVirtualization) CreateVM() error {
	if v.connection == nil {
		return errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	domainDefinition, err := v.connection.DomainDefineXML(`
        <domain type='kvm'>
          <name>test_domain</name>
          <memory unit='MB'>2048</memory>
          <vcpu>4</vcpu>

          <os>
            <type>hvm</type>
          </os>

          <devices>
            <disk type='file' device='disk'>
              <driver name='qemu' type='qcow2'/>
              <source file='/var/lib/libvirt/images/test.img'/>
              <target dev='vda' bus='virtio'/>
            </disk>
            <disk type='file' device='cdrom'>
              <source file='/var/lib/libvirt/boot/ubuntu-server.iso'/>
              <target dev='hdc' bus='ide'/>
              <readonly/>
            </disk>
            <interface type='network'>
              <source network='default'/>
            </interface>
            <serial type='pty'>
              <target port='0'/>
            </serial>
            <console type='pty'>
              <target type='serial' port='0'/>
            </console>
          </devices>

          <on_reboot>restart</on_reboot>
          <on_crash>restart</on_crash>
        </domain>
        `)
	if err != nil {
		return err
	}

	err = v.connection.DomainCreate(domainDefinition)
	if err != nil {
		return err
	}

	return nil
}

func (v KvmVirtualization) CreateVolume() error {
	if v.connection == nil {
		return errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	storagePools, err := v.ListStoragePools()
	if err != nil {
		return err
	}

	storageVolCreateFlags := libvirt.StorageVolCreateFlags(0)
	_, err = v.connection.StorageVolCreateXML(storagePools[0],
		`
        <volume type='file'>
          <name>test.img</name>
          <capacity unit="MB">2000</capacity>
          <target>
            <format type='qcow2'/>
            <permissions>
              <mode>0755</mode>
            </permissions>
          </target>
        </volume>
        `, storageVolCreateFlags)
	if err != nil {
		return err
	}

	return nil
}

func (v KvmVirtualization) CreatePool() error {
	if v.connection == nil {
		return errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	storagePools, err := v.ListStoragePools()
	if err != nil {
		return err
	}

	// check if storage pools is already created
	if len(storagePools) != 0 {
		slog.Info("storage pool is already created")
		slog.Info("checking if storage pool is started...")

		isActive, err := v.connection.StoragePoolIsActive(storagePools[0])
		if isActive == -1 { // indicates an error
			return errors.New(LIBVIRT_POOL_ERROR)
		}
		if isActive == 0 { // inactive
			storagePoolCreateFlags := libvirt.StoragePoolCreateNormal
			err = v.connection.StoragePoolCreate(
				storagePools[0],
				storagePoolCreateFlags,
			)
			if err != nil {
				return err
			}
		}

		slog.Info("storage pool is running")

		return nil
	}

	createStoragePoolsFlags := libvirt.StoragePoolCreateNormal
	_, err = v.connection.StoragePoolCreateXML(
		`
        <pool type="dir">
          <name>vm_storage_pool</name>
          <target>
            <path>/var/lib/libvirt/images</path>
          </target>
        </pool>
        `, createStoragePoolsFlags)
	if err != nil {
		return err
	}

	return nil
}

func (v KvmVirtualization) ListStoragePools() ([]libvirt.StoragePool, error) {
	if v.connection == nil {
		return nil, errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	// for now, only list the storage pools
	// that already defined and either online or offline
	listStoragePoolsFlags := libvirt.ConnectListStoragePoolsActive
	listStoragePoolsFlags |= libvirt.ConnectListStoragePoolsInactive
	storagePools, _, err := v.connection.ConnectListAllStoragePools(
		1,
		listStoragePoolsFlags,
	)
	if err != nil {
		return nil, err
	}

	return storagePools, nil
}
