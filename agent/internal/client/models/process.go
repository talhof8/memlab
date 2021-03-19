package models

import (
	"github.com/memlab/agent/internal/types"
	psUtil "github.com/shirou/gopsutil/process"
	"gopkg.in/guregu/null.v3"
)

type Process struct {
	ID             string    `json:"id,omitempty"`
	Pid            types.Pid `json:"pid"`
	Executable     string    `json:"executable"`
	CommandLine    string    `json:"command_line"`
	CreateTime     null.Time `json:"create_time"`
	LastSeenAt     null.Time `json:"last_seen_at"`
	Monitored      bool      `json:"monitored,omitempty"`
	MonitoredSince null.Time `json:"monitored_since,omitempty"`
	Status         string    `json:"status"`
}

func (p *Process) LiveProcess() (*psUtil.Process, error) {
	return psUtil.NewProcess(int32(p.Pid.Uint32()))
}
