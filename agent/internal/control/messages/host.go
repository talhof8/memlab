package messages

import (
	"github.com/glendc/go-external-ip"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/host"
)

var (
	ipAddressResolver = externalip.DefaultConsensus(nil, nil)
)

type HostStatusReport struct {
	MachineId            string `json:"machine_id"`
	PublicIpAddress      string `json:"public_ip_address"`
	Hostname             string `json:"hostname"`
	LastBootTime         uint64 `json:"last_boot_at"`
	OS                   string `json:"operating_system"`
	Platform             string `json:"platform"`
	PlatformFamily       string `json:"platform_family"`
	PlatformVersion      string `json:"platform_version"`
	KernelVersion        string `json:"kernel_version"`
	KernelArch           string `json:"kernel_architecture"`
	VirtualizationSystem string `json:"virtualization_system"`
	VirtualizationRole   string `json:"virtualization_role"`
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
	hostStatusReport.LastBootTime = hostInfo.BootTime
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
