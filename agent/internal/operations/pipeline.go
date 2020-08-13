package operations

import (
	"context"
	"github.com/memlab/agent/internal/operations/operators"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"sync"
)

type Pipeline struct {
	logger    *zap.Logger
	waitGroup sync.WaitGroup
	context   context.Context
	cancel    context.CancelFunc
	operators []operators.Operator
	running   *atomic.Bool
}

func NewPipeline(ctx context.Context, rootLogger *zap.Logger) *Pipeline {
	logger := rootLogger.Named("operations-pipeline")
	ctx, cancel := context.WithCancel(ctx)

	return &Pipeline{
		logger:    logger,
		context:   ctx,
		cancel:    cancel,
		operators: make([]operators.Operator, 0),
		running:   atomic.NewBool(false),
	}
}

func (p *Pipeline) AddOperators(ops ...operators.Operator) {
	p.operators = append(p.operators, ops...)
}

func (p *Pipeline) Start() error {
	p.running.Toggle() // Turn on
	return nil
}

func (p *Pipeline) WaitUntilCompletion() {
	p.waitGroup.Wait()
	p.running.Toggle() // Turn off
}

func (p *Pipeline) Running() bool {
	return p.running.Load()
}

func (p *Pipeline) Stop() error {
	p.cancel()
	return nil
}
