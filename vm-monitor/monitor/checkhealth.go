package monitor

import (
	"context"
	"log/slog"
	"ta-vm-monitor/model"

	"libvirt.org/go/libvirt"
)

type CheckhealthMonitor struct {
	libvirtConn    *libvirt.Connect
	nodeRepository *NodeRepository
	vmMap          map[string]bool
}

func NewCheckhealthMonitor(
	libvirtConn *libvirt.Connect,
	nodeRepository *NodeRepository,
) *CheckhealthMonitor {
	return &CheckhealthMonitor{
		libvirtConn:    libvirtConn,
		nodeRepository: nodeRepository,
		vmMap:          make(map[string]bool),
	}
}

func (m *CheckhealthMonitor) GetDomain(
	domainName string,
) (*libvirt.Domain, error) {
	dom, err := m.libvirtConn.LookupDomainByName(domainName)
	if err != nil {
		return nil, err
	}

	return dom, nil
}

func (m *CheckhealthMonitor) GetAllDomains() ([]libvirt.Domain, error) {
	activeAndInactiveFlags := 0 |
		libvirt.CONNECT_LIST_DOMAINS_ACTIVE |
		libvirt.CONNECT_LIST_DOMAINS_INACTIVE

	return m.libvirtConn.ListAllDomains(activeAndInactiveFlags)
}

func (m *CheckhealthMonitor) DomainsCheck() ([]model.NodeStatus, error) {
	allDomainStatus := []model.NodeStatus{}

	allDomain, err := m.nodeRepository.GetAllNodes(context.Background())
	if err != nil {
		return allDomainStatus, err
	}

	for _, domain := range allDomain {
		thisDomainStatus := new(model.NodeStatus)

		thisDomainStatus.Name = domain.Name

		vmDomain, err := m.libvirtConn.LookupDomainByName(domain.Name)
		if err != nil {
			thisDomainStatus.Error = err
			continue
		}

		isDomainActive, err := vmDomain.IsActive()
		if err != nil {
			thisDomainStatus.Error = err
			continue
		}
		thisDomainStatus.IsActive = isDomainActive

		thisDomainCpuStats, err := vmDomain.GetCPUStats(-1, 1, 0)
		if err != nil {
			thisDomainStatus.Error = err
			continue
		}
		thisDomainStatus.CpuTime = thisDomainCpuStats[0].CpuTime

		allDomainStatus = append(allDomainStatus, *thisDomainStatus)
	}

	return allDomainStatus, nil
}

func (m *CheckhealthMonitor) Restart(domain libvirt.Domain) {
	domainIsActive, err := domain.IsActive()
	if err != nil {
		slog.Error("could not get domain status")
		return
	}

	if domainIsActive {
		return
	}

	err = domain.Reboot(libvirt.DOMAIN_REBOOT_DEFAULT)
	if err != nil {
		slog.Error("could not reboot domain")
		return
	}
}
