package operations

import (
	"context"
	"github.com/memlab/agent/internal/operations/operators"
	"github.com/memlab/agent/internal/reports"
	"github.com/memlab/agent/internal/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

const defaultOperatorContextTimeout = time.Minute

var ErrOperatorFailure = errors.New("operator failure")

type Pipeline struct {
	logger    *zap.Logger
	context   context.Context
	cancel    context.CancelFunc
	operators []operators.Operator
}

func NewPipeline(ctx context.Context, rootLogger *zap.Logger, ops []operators.Operator) *Pipeline {
	logger := rootLogger.Named("operations-pipeline")
	ctx, cancel := context.WithCancel(ctx)

	return &Pipeline{
		logger:    logger,
		context:   ctx,
		cancel:    cancel,
		operators: ops,
	}
}

func (p *Pipeline) AddOperators(ops ...operators.Operator) {
	p.operators = append(p.operators, ops...)
}

func (p *Pipeline) Run(pid types.Pid) (map[string]interface{}, error) {
	mergedReportsDump, err := p.runOperators(pid)
	if err != nil {
		return nil, err
	}

	return mergedReportsDump, nil
}

func (p *Pipeline) runOperators(pid types.Pid) (map[string]interface{}, error) {
	allReports := make([]reports.Report, 0)

	for _, operator := range p.operators {
		operatorContext, cancelOperator := context.WithTimeout(p.context, defaultOperatorContextTimeout)

		report, err := operator.Operate(operatorContext, pid)
		if err != nil {
			cancelOperator()

			p.logger.Error("Operator failed", zap.String("OperatorName", operator.OperatorName()),
				zap.Bool("FailPipelineOnError", operator.FailPipelineOnError()), zap.Error(err))

			if operator.FailPipelineOnError() {
				return nil, ErrOperatorFailure
			}
		}

		allReports = append(allReports, report)
	}

	return reports.MergeReports(allReports...)
}

func (p *Pipeline) Abort() error {
	p.cancel()
	return nil
}
