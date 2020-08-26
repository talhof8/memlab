package operators

import (
	"context"
	"github.com/memlab/agent/internal/reports"
	"github.com/memlab/agent/internal/reports/postdetection"
	"github.com/memlab/agent/internal/types"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

type ProcDumpOperator struct{}

func (p *ProcDumpOperator) OperatorName() string {
	return "proc-dump-operator"
}

func (p *ProcDumpOperator) Operate(ctx context.Context, pid types.Pid) (reports.Report, error) {
	// 1. take proc dump
	// 2. start a goroutine which uploads it (and don't wait for it to finish? perhaps dump it in a background pool
	// which uploads to backend/other destination and keeps internal state persistently in case it didn't finish)

	ps, err := process.NewProcess(int32(pid))
	if err != nil {
		if errors.Cause(err) == process.ErrorProcessNotRunning {
			return nil, errors.Errorf("process '%d' is not running", pid)
		}

		return nil, err
	}
	/**
	cmd := exec.Command("sudo", "procdump", "-p", fmt.Sprintf("%d", caughtSignal.Pid))
	output, err := cmd.Output()
	if err != nil {
		funcLogger.Error("Failed to run command", zap.Error(err))
		return
	}

	*/

	return postdetection.NewMetadataReport(ps)
}

func (p *ProcDumpOperator) FailPipelineOnError() bool {
	return false
}
