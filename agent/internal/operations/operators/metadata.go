package operators

import (
	"context"
	"github.com/memlab/agent/internal/reports"
	"github.com/memlab/agent/internal/reports/postdetection"
	"github.com/memlab/agent/internal/types"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

type CollectMetadata struct{}

func (c *CollectMetadata) OperatorName() string {
	return "collect-metadata-operator"
}

func (c *CollectMetadata) Operate(ctx context.Context, pid types.Pid) (reports.Report, error) {
	ps, err := process.NewProcess(int32(pid))
	if err != nil {
		if errors.Cause(err) == process.ErrorProcessNotRunning {
			return nil, errors.Errorf("process '%d' is not running", pid)
		}

		return nil, err
	}

	return postdetection.NewMetadataReport(ps)
}

func (c *CollectMetadata) FailPipelineOnError() bool {
	return false
}
