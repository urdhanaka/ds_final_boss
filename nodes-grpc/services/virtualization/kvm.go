package virtualization

import (
	"errors"
	"log/slog"

	"github.com/digitalocean/go-libvirt"
)

const (
	// cpu threshold
	NODE_BUSY_THRESHOLD float64 = 60
)

type KvmVirtualization struct {
	connection  *libvirt.Libvirt
	IsReady     bool
	Utilization float64
}

func NewKvmVirt() *KvmVirtualization {
	return &KvmVirtualization{
		connection:  initKvmConnection(),
		IsReady:     false,
		Utilization: float64(0),
	}
}

func (v KvmVirtualization) CreatePool() error {
	if v.connection == nil {
		return errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	listStoragePoolsFlags := libvirt.ConnectListStoragePoolsActive
	listStoragePoolsFlags |= libvirt.ConnectListStoragePoolsInactive
	storagePools, _, err := v.connection.ConnectListAllStoragePools(1, listStoragePoolsFlags)
	if err != nil {
		return err
	}

	// check if storage pools is already created
	if len(storagePools) != 0 {
		slog.Info("storage pool is already created")
		return nil
	}

	createStoragePoolsFlags := libvirt.StoragePoolCreateNormal
	_, err = v.connection.StoragePoolCreateXML(
		`
        <pool type="dir">
          <name>vm_storage_pool</name>
          <target>
            <path>/home/rpl-02/UrdhanakaAptanagi</path>
          </target>
        </pool>
        `, createStoragePoolsFlags)
	if err != nil {
		return err
	}

	return nil
}

func (v KvmVirtualization) CreateVM() error {
    if v.connection == nil {
		return errors.New(LIBVIRT_NOT_INITIALIZED)
    }

    _, err := v.connection.DomainDefineXML(`
        <domain type='kvm'>
          <name>test_domain</name>
          <memory unit='MB'>2048</memory>
          <vcpu>4</vcpu>
          <os>
            <type>hvm</type>
          </os>
          <devices>
          </devices>
          <on_reboot>restart</on_reboot>
          <on_crash>restart</on_crash>
        </domain>
        `)
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
    
    storageVolCreateFlags := libvirt.StorageVolCreatePreallocMetadata
    _, err = v.connection.StorageVolCreateXML(storagePools[0], 
        `
        <volume>
          <name>test</name>
          <allocation unit="MB">512</allocation>
          <capacity unit="MB">512</capacity>
          <target>
            <format type='qcow2'/>
            <permissions>
              <label>virt_image_t</label>
            </permissions>
          </target>
        </volume>
        `, storageVolCreateFlags)
    if err != nil {
        return err
    }

	return nil
}

func (v KvmVirtualization) ListStoragePools() ([]libvirt.StoragePool, error) {
	if v.connection == nil {
		return nil, errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	listStoragePoolsFlags := libvirt.ConnectListStoragePoolsActive
	listStoragePoolsFlags |= libvirt.ConnectListStoragePoolsInactive
	storagePools, _, err := v.connection.ConnectListAllStoragePools(1, listStoragePoolsFlags)
	if err != nil {
		return nil, err
	}

	return storagePools, nil
}
