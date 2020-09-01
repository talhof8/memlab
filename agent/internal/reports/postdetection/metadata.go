package postdetection

import (
	"context"
	"encoding/json"
	"github.com/memlab/agent/internal/host"
	"github.com/memlab/agent/internal/types"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

// todo: get host's top 10 consuming mem and cpu process, and general cpu and mem utilization

const maxConnectionsLimit = 50

type MetadataReport struct {
	Pid            types.Pid `json:"pid"`
	MachineId      string    `json:"machine_id"`
	ExecutablePath string    `json:"executable_path"`
	CmdLine        string    `json:"cmd_line"`
	CpuPercent     float64   `json:"cpu_percent"`
	MemPercent     float32   `json:"memory_percent"`
	CreateTime     int64     `json:"create_time"`
	Cwd            string    `json:"cwd"`
	Connections    []string  `json:"connections"`
}

func NewMetadataReport(ctx context.Context, pid types.Pid, ps *process.Process) (*MetadataReport, error) {
	machineId, err := host.MachineId()
	if err != nil {
		return nil, err
	}

	executablePath, err := ps.ExeWithContext(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "get process' executable (pid: '%d')", pid)
	}

	cmdline, err := ps.CmdlineWithContext(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "get process' cmdline (pid: '%d')", pid)
	}

	cpuPercent, err := ps.CPUPercentWithContext(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "get process' CPU percent (pid: '%d')", pid)
	}

	memPercent, err := ps.MemoryPercentWithContext(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "get process' memory percent (pid: '%d')", pid)
	}

	createTime, err := ps.CreateTimeWithContext(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "get process' create time (pid: '%d')", pid)
	}

	cwd, err := ps.CwdWithContext(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "get process' cwd (pid: '%d')", pid)
	}

	connections, err := listConnections(ctx, ps)
	if err != nil {
		return nil, errors.WithMessagef(err, "get process' connections (pid: '%d')", pid)
	}

	return &MetadataReport{
		MachineId:      machineId,
		Pid:            pid,
		ExecutablePath: executablePath,
		CmdLine:        cmdline,
		CpuPercent:     cpuPercent,
		MemPercent:     memPercent,
		CreateTime:     createTime,
		Cwd:            cwd,
		Connections:    connections,
	}, nil
}

func listConnections(ctx context.Context, ps *process.Process) ([]string, error) {
	rawConnectionList, err := ps.ConnectionsMaxWithContext(ctx, maxConnectionsLimit)
	if err != nil {
		return nil, err
	}

	connections := make([]string, 0, len(rawConnectionList))
	for _, rawConnection := range rawConnectionList {
		connections = append(connections, rawConnection.String())
	}
	return connections, nil
}

func (m *MetadataReport) ReportName() string {
	return "metadata-report"
}

func (m *MetadataReport) DumpReport() ([]byte, error) {
	return json.Marshal(m)
}
