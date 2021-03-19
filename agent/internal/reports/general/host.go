package general

import (
	"encoding/json"
	"github.com/glendc/go-external-ip"
	"github.com/memlab/agent/internal/client/models"
	"github.com/memlab/agent/internal/types"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/host"
)

var (
	ipAddressResolver = externalip.DefaultConsensus(nil, nil)
)

type HostStatusReport struct {
	*models.Host
}

func NewHostStatusReport(machineId string) (*HostStatusReport, error) {
	hostStatusReport := &HostStatusReport{}

	hostStatusReport.MachineId = machineId

	hostInfo, err := host.Info()
	if err != nil {
		return nil, errors.WithMessage(err, "get host info")
	}

	// todo: add a cache which self-updates every once in a while to save redundant outgoing traffic.
	publicIpAddress, err := ipAddressResolver.ExternalIP()
	if err != nil {
		return nil, errors.WithMessage(err, "get external ip address")
	}
	hostStatusReport.PublicIpAddress = publicIpAddress.String()

	hostStatusReport.Hostname = hostInfo.Hostname
	hostStatusReport.LastBootTime = types.JsonTimeFromTimestamp(int64(hostInfo.BootTime))
	hostStatusReport.OS = hostInfo.OS
	hostStatusReport.Platform = hostInfo.Platform
	hostStatusReport.PlatformFamily = hostInfo.PlatformFamily
	hostStatusReport.PlatformVersion = hostInfo.PlatformVersion
	hostStatusReport.KernelVersion = hostInfo.KernelVersion
	hostStatusReport.KernelArch = hostInfo.KernelArch
	hostStatusReport.VirtualizationSystem = hostInfo.VirtualizationSystem
	hostStatusReport.VirtualizationRole = hostInfo.VirtualizationRole

	return hostStatusReport, nil
}

func (h *HostStatusReport) ReportName() string {
	return "host-status-report"
}

func (h *HostStatusReport) DumpReport() ([]byte, error) {
	return json.Marshal(h)
}
