package virtualization

import (
	"errors"
	"fmt"
	"log/slog"
	virtualization_model "nodes-grpc/common/model/virtualization"
	"nodes-grpc/services/db"
	"os"
	"strings"

	"github.com/digitalocean/go-libvirt"
	"github.com/google/uuid"
)

const (
	// cpu threshold
	NODE_BUSY_THRESHOLD float64 = 60

	// libvirt-related error
	LIBVIRT_NOT_INITIALIZED string = "libvirt connection is not initialized"

	// storage pool name
	STORAGE_POOL_NAME string = "default"
)

type LibvirtVirtualization struct {
	libvirtConnection *libvirt.Libvirt
	dbConnection      db.DatabaseInterface
}

func NewLibvirtService(
	libvirtConnection *libvirt.Libvirt, dbConnection db.DatabaseInterface,
) *LibvirtVirtualization {
	return &LibvirtVirtualization{
		libvirtConnection: libvirtConnection,
		dbConnection:      dbConnection,
	}
}

func (v *LibvirtVirtualization) Spawn(
	virtRequest virtualization_model.NodeCreateRequest,
) error {
	slog.Info("Spawn(): creating vm instance using libvirt...")

	if v.libvirtConnection == nil {
		return errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	_, err := v.CreateVolume(virtualization_model.VirtualizationCreateRequest{})
	if err != nil {
		slog.Error("Spawn(): error creating volume",
			"error", err.Error(),
		)
		return err
	}
	v.CreateVM(virtualization_model.VirtualizationCreateRequest{})

	return nil
}

func (v *LibvirtVirtualization) Stop(virtRequest virtualization_model.InstanceIdentification) error {
	return nil
}

func (v *LibvirtVirtualization) List() error {
	return nil
}

func (v LibvirtVirtualization) CreateVM(
	virtRequest virtualization_model.VirtualizationCreateRequest,
) error {
	if v.libvirtConnection == nil {
		return errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	template, err := os.ReadFile("./common/templates/domain-template.xml")
	if err != nil {
		slog.Error("CreateVM(): could not get template file",
			"error", err.Error(),
		)
		return err
	}

	domainDefinition, err := v.libvirtConnection.DomainDefineXML(string(template))
	if err != nil {
		return err
	}

	err = v.libvirtConnection.DomainCreate(domainDefinition)
	if err != nil {
		return err
	}

	return nil
}

func (v LibvirtVirtualization) CreateVolume(
	virtRequest virtualization_model.VirtualizationCreateRequest,
) (string, error) {
	if v.libvirtConnection == nil {
		return "", errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	storagePool, err := v.libvirtConnection.StoragePoolLookupByName(STORAGE_POOL_NAME)
	if err != nil {
		return "", err
	}

	volumeName := uuid.New().String()

	storageVolCreateFlags := libvirt.StorageVolCreateFlags(1)
	_, err = v.libvirtConnection.StorageVolCreateXML(storagePool,
		fmt.Sprintf(`
            <volume type='file'>
                <name>test.qcow2</name>
                <key>/var/lib/libvirt/images/test.qcow2</key>
                <capacity unit='bytes'>4294967296</capacity>
                <allocation unit='bytes'>856064</allocation>
                <physical unit='bytes'>4295884800</physical>
                <target>
                    <path>/var/lib/libvirt/images/test.qcow2</path>
                    <format type='qcow2'/>
                    <permissions>
                        <mode>0600</mode>
                        <owner>107</owner>
                        <group>107</group>
                        <label>system_u:object_r:svirt_image_t:s0:c723,c1015</label>
                    </permissions>
                    <compat>1.1</compat>
                    <clusterSize unit='B'>65536</clusterSize>
                    <features>
                        <lazy_refcounts/>
                    </features>
                </target>
            </volume>
        `), storageVolCreateFlags)
	if err != nil {
		return "", err
	}

	return volumeName, nil
}

func (v LibvirtVirtualization) CreatePool() error {
	if v.libvirtConnection == nil {
		return errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	storagePool, err := v.libvirtConnection.StoragePoolLookupByName(STORAGE_POOL_NAME)
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
                    `, STORAGE_POOL_NAME), libvirt.StoragePoolCreateNormal)
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

func (v LibvirtVirtualization) CreateNetwork() error {
    v.libvirtConnection.NetworkCreateXML()

	return nil
}
