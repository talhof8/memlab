package postdetection

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

// todo: get host's top 10 consuming mem and cpu process, and general cpu and mem utilization

type MetadataReport struct {
	ExecutablePath string  `json:"executable_path"`
	CmdLine        string  `json:"cmd_line"`
	CpuPercent     float64 `json:"cpu_percent"`
	MemPercent     float32 `json:"memory_percent"`
	CreateTime     int64   `json:"create_time"`
	Cwd            string  `json:"cwd"`
}

func NewMetadataReport(ps *process.Process) (*MetadataReport, error) {
	pid := ps.Pid

	executablePath, err := ps.Exe()
	if err != nil {
		return nil, errors.WithMessagef(err, "get process' executable (pid: '%d')", pid)
	}

	cmdline, err := ps.Cmdline()
	if err != nil {
		return nil, errors.WithMessagef(err, "get process' cmdline (pid: '%d')", pid)
	}

	cpuPercent, err := ps.CPUPercent()
	if err != nil {
		return nil, errors.WithMessagef(err, "get process' CPU percent (pid: '%d')", pid)
	}

	memPercent, err := ps.MemoryPercent()
	if err != nil {
		return nil, errors.WithMessagef(err, "get process' memory percent (pid: '%d')", pid)
	}

	createTime, err := ps.CreateTime()
	if err != nil {
		return nil, errors.WithMessagef(err, "get process' create time (pid: '%d')", pid)
	}

	cwd, err := ps.Cwd()
	if err != nil {
		return nil, errors.WithMessagef(err, "get process' cwd (pid: '%d')", pid)
	}

	return &MetadataReport{
		ExecutablePath: executablePath,
		CmdLine:        cmdline,
		CpuPercent:     cpuPercent,
		MemPercent:     memPercent,
		CreateTime:     createTime,
		Cwd:            cwd,
	}, nil
}

func (m *MetadataReport) ReportName() string {
	return "metadata-report"
}

func (m *MetadataReport) DumpReport() ([]byte, error) {
	return json.Marshal(m)
}
