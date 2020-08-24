package messages

import (
	"github.com/memlab/agent/internal/control/responses"
	"github.com/pkg/errors"
	psUtil "github.com/shirou/gopsutil/process"
	"time"
)

type SingleProcessReport struct {
	RecordId    string     `json:"id"`
	MachineId   string     `json:"machine_id"`
	Pid         int32      `json:"pid"`
	Executable  string     `json:"executable"`
	CommandLine string     `json:"command_line"`
	CreateTime  *time.Time `json:"create_time"`
	LastSeenAt  *time.Time `json:"last_seen_at"`
	Status      string     `json:"status"`
}

type ProcessListReport = []*SingleProcessReport

func NewProcessListReport(machineId string, backendProcesses map[int32]*responses.Process) (*ProcessListReport, error) {
	liveProcesses, err := psUtil.Processes()
	if err != nil {
		return nil, errors.WithMessage(err, "get live process list")
	}

	report := make(ProcessListReport, 0, len(liveProcesses))

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

		singleProcessReport := &SingleProcessReport{
			RecordId:    "",
			MachineId:   machineId,
			Pid:         liveProcess.Pid,
			Executable:  executable,
			CommandLine: cmdLine,
			CreateTime:  &createTime,
			LastSeenAt:  &now,
			Status:      status,
		}

		backendProcess, matched := backendProcesses[liveProcess.Pid]

		// In case this machine's process is already listed in backend, meaning both same pid and same creation time.
		if matched && machineId == backendProcess.MachineId && createTime.Equal(*backendProcess.CreateTime) {
			singleProcessReport.RecordId = backendProcess.RecordId
		}

		// A whole new process or a new process with same pid
		report = append(report, singleProcessReport)
	}

	return report, nil
}
