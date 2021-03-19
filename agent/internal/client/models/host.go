package models

import "gopkg.in/guregu/null.v3"

type Host struct {
	ID                   string    `json:"id,omitempty"`
	MachineId            string    `json:"machine_id"`
	PublicIpAddress      string    `json:"public_ip_address"`
	Hostname             string    `json:"hostname"`
	LastBootTime         null.Time `json:"last_boot_at"`
	OS                   string    `json:"operating_system"`
	Platform             string    `json:"platform"`
	PlatformFamily       string    `json:"platform_family"`
	PlatformVersion      string    `json:"platform_version"`
	KernelVersion        string    `json:"kernel_version"`
	KernelArch           string    `json:"kernel_architecture"`
	VirtualizationSystem string    `json:"virtualization_system"`
	VirtualizationRole   string    `json:"virtualization_role"`
}
