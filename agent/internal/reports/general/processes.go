package general

import (
	"encoding/json"
	"github.com/hashicorp/go-multierror"
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
	MachineId string         `json:"machine_id"`
	List      []*HostProcess `json:"processes"`
}

func NewProcessListReport(machineId string) (*ProcessListReport, error) {
	liveProcesses, err := psUtil.Processes()
	if err != nil {
		return nil, errors.WithMessage(err, "get live process list")
	}

	hostProcesses := make([]*HostProcess, 0, len(liveProcesses))

	var errs error

	for _, liveProcess := range liveProcesses {
		executable, err := liveProcess.Exe()
		if err != nil {
			errs = multierror.Append(errs, errors.WithMessagef(err, "get executable for pid '%d'", liveProcess.Pid))
			continue
		}

		cmdLine, err := liveProcess.Cmdline()
		if err != nil {
			errs = multierror.Append(errs, errors.WithMessagef(err, "get command line for pid '%d'", liveProcess.Pid))
			continue
		}

		createTimeMilliseconds, err := liveProcess.CreateTime()
		if err != nil {
			errs = multierror.Append(errs, errors.WithMessagef(err, "get create time for pid '%d'", liveProcess.Pid))
			continue
		}
		createTime := types.TimeFromMillisecondTimestamp(createTimeMilliseconds)
		now := time.Now().UTC()

		status, err := liveProcess.Status()
		if err != nil {
			errs = multierror.Append(errs, errors.WithMessagef(err, "get status for pid '%d'", liveProcess.Pid))
			continue
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

	// todo: currently missing out errors if host processes' length is not 0, at the very least log them instead.

	if len(hostProcesses) == 0 && errs != nil {
		return nil, errs
	}

	return &ProcessListReport{MachineId: machineId, List: hostProcesses}, nil
}

func (p *ProcessListReport) ReportName() string {
	return "process-list-report"
}

func (p *ProcessListReport) DumpReport() ([]byte, error) {
	return json.Marshal(p)
}
