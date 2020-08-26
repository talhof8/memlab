package operations

import (
	"context"
	"github.com/memlab/agent/internal/operations/operators"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var ErrOperatorFailure = errors.New("operator failure")

type Pipeline struct {
	logger    *zap.Logger
	context   context.Context
	cancel    context.CancelFunc
	operators []operators.Operator
}

func NewPipeline(ctx context.Context, rootLogger *zap.Logger) *Pipeline {
	logger := rootLogger.Named("operations-pipeline")
	ctx, cancel := context.WithCancel(ctx)

	return &Pipeline{
		logger:    logger,
		context:   ctx,
		cancel:    cancel,
		operators: make([]operators.Operator, 0),
	}
}

func (p *Pipeline) AddOperators(ops ...operators.Operator) {
	p.operators = append(p.operators, ops...)
}

func (p *Pipeline) Start() error {
	if err := p.runOperators(); err != nil {
		return err
	}

	return nil
}

func (p *Pipeline) runOperators() error {
	for _, operator := range p.operators {
		err := operator.Operate(p.context)
		if err != nil {
			p.logger.Error("Operator failed", zap.String("OperatorName", operator.OperatorName()),
				zap.Bool("FailPipelineOnError", operator.FailPipelineOnError()), zap.Error(err))

			if operator.FailPipelineOnError() {
				return ErrOperatorFailure
			}
		}
	}

	return nil
}

func (p *Pipeline) Stop() error {
	p.cancel()
	return nil
}
