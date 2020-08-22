package control

import (
	"context"
	"github.com/memlab/agent/internal/control/client"
	"go.uber.org/zap"
	"sync"
)

type Plane struct {
	logger    *zap.Logger
	context   context.Context
	cancel    context.CancelFunc
	waitGroup sync.WaitGroup
	client    *client.RestfulClient
}

func NewPlane(ctx context.Context, rootLogger *zap.Logger, apiConfig *client.ApiConfig) (*Plane, error) {
	logger := rootLogger.Named("control-plane")
	ctx, cancel := context.WithCancel(ctx)

}

func (p *Plane) Start() error                      {}
func (p *Plane) Stop() error                       {}
func (p *Plane) SendProcessList() error            {}
func (p *Plane) SendHostStatus() error             {}
func (p *Plane) SubscribeToMonitorCommands() error {}
func (p *Plane) SendProcessStatus() error          {}
