package postdetection

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

type ProcDumpReport struct {
	Size uint64
}

func NewProcDumpReport(ps *process.Process) (*ProcDumpReport, error) {
	pid := ps.Pid

	return &ProcDumpReport{}, nil
}

func (m *ProcDumpReport) ReportName() string {
	return "proc-dump-report"
}

func (m *ProcDumpReport) DumpReport() ([]byte, error) {
	return json.Marshal(m)
}
