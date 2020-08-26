package operators

import (
	"context"
	"github.com/memlab/agent/internal/reports"
	"github.com/memlab/agent/internal/types"
)

type Operator interface {
	OperatorName() string
	Operate(ctx context.Context, pid types.Pid) (reports.Report, error)
	FailPipelineOnError() bool
}
