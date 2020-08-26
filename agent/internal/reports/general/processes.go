package general

import (
	"encoding/json"
	"github.com/memlab/agent/internal/types"
	"github.com/pkg/errors"
	psUtil "github.com/shirou/gopsutil/process"
	"time"
)

type HostProcess struct {
	Pid         types.Pid `json:"pid"`
	Executable  string    `json:"executable"`
	CommandLine string    `json:"command_line"`
	CreateTime  time.Time `json:"create_time"`
	LastSeenAt  time.Time `json:"last_seen_at"`
	Status      string    `json:"status"`
}

type ProcessListReport struct {
	list []*HostProcess
}

func NewProcessListReport() (*ProcessListReport, error) {
	liveProcesses, err := psUtil.Processes()
	if err != nil {
		return nil, errors.WithMessage(err, "get live process list")
	}

	hostProcesses := make([]*HostProcess, 0, len(liveProcesses))

	for _, liveProcess := range liveProcesses {
		executable, err := liveProcess.Exe()
		if err != nil {
			return nil, errors.WithMessagef(err, "get executable for pid '%d'", liveProcess.Pid)
		}

		cmdLine, err := liveProcess.Cmdline()
		if err != nil {
			return nil, errors.WithMessagef(err, "get command line for pid '%d'", liveProcess.Pid)
		}

		createTimeMilliseconds, err := liveProcess.CreateTime()
		if err != nil {
			return nil, errors.WithMessagef(err, "get create time for pid '%d'", liveProcess.Pid)
		}
		createTime := time.Unix(createTimeMilliseconds, 0).UTC()
		now := time.Now().UTC()

		status, err := liveProcess.Status()
		if err != nil {
			return nil, errors.WithMessagef(err, "get status for pid '%d'", liveProcess.Pid)
		}

		hostProcess := &HostProcess{
			Pid:         types.Pid(liveProcess.Pid),
			Executable:  executable,
			CommandLine: cmdLine,
			CreateTime:  createTime,
			LastSeenAt:  now,
			Status:      status,
		}

		hostProcesses = append(hostProcesses, hostProcess)
	}

	return &ProcessListReport{list: hostProcesses}, nil
}

func (p *ProcessListReport) ReportName() string {
	return "process-list-report"
}

func (p *ProcessListReport) DumpReport() ([]byte, error) {
	return json.Marshal(p.list)
}
