package control

import (
	"context"
	"github.com/memlab/agent/internal/control/client"
	"github.com/memlab/agent/internal/control/messages"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

const (
	endpointHosts     = "hosts"
	endpointProcesses = "processes"
)

type PlaneConfig struct {
	ApiConfig                 *client.ApiConfig
	HostStatusReportInterval  time.Duration
	ProcessListReportInterval time.Duration
}

func (pc *PlaneConfig) Valid() (bool, error) {
	if pc.HostStatusReportInterval <= 0 {
		return false, errors.New("uninitialized host status report interval")
	} else if pc.ProcessListReportInterval <= 0 {
		return false, errors.New("uninitialized process list report interval")
	}

	return true, nil
}

type Plane struct {
	logger    *zap.Logger
	context   context.Context
	cancel    context.CancelFunc
	waitGroup sync.WaitGroup
	client    *client.RestfulClient
	config    *PlaneConfig
}

func NewPlane(ctx context.Context, rootLogger *zap.Logger, config *PlaneConfig) (*Plane, error) {
	if valid, err := config.Valid(); !valid {
		return nil, errors.WithMessage(err, "validate control plane config")
	}

	logger := rootLogger.Named("control-plane")
	ctx, cancel := context.WithCancel(ctx)
	restfulClient, err := client.NewRestfulClient(ctx, logger, config.ApiConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "new restful client")
	}

	return &Plane{
		context: ctx,
		cancel:  cancel,
		client:  restfulClient,
		config:  config,
	}, nil
}

func (p *Plane) Start() error {
	p.logger.Debug("Start control plane")

	p.waitGroup.Add(1)
	go p.reportHostStatus()

	p.waitGroup.Add(1)
	go p.reportProcessList()

	p.waitGroup.Add(1)
	go p.subscribeToMonitorCommands()

	return nil
}

func (p *Plane) reportHostStatus() {
	defer p.waitGroup.Done()

	ticker := time.NewTicker(p.config.HostStatusReportInterval)
	for {
		select {
		case <-p.context.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			message, err := messages.NewHostStatusReport()
			if err != nil {
				p.logger.Error("Failed to create host status report", zap.Error(err))
				continue
			}

			response, err := p.client.Post(endpointHosts, message)
			if err != nil {
				p.logger.Error("Failed to send host status report", zap.Error(err))
				continue
			} else if response.StatusCode != http.StatusCreated {
				p.logger.Warn("Got a bad status code", zap.Int("Got", response.StatusCode),
					zap.Int("Expected", http.StatusOK))
				continue
			}
		}
	}
}

func (p *Plane) reportProcessList() {
	defer p.waitGroup.Done()

	ticker := time.NewTicker(p.config.ProcessListReportInterval)
	for {
		select {
		case <-p.context.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			message, err := messages.NewProcessListReport()
			if err != nil {
				p.logger.Error("Failed to create process list report", zap.Error(err))
				continue
			}

			response, err := p.client.Post(endpointProcesses, message)
			if err != nil {
				p.logger.Error("Failed to send process list report", zap.Error(err))
				continue
			} else if response.StatusCode != http.StatusCreated {
				p.logger.Warn("Got a bad status code", zap.Int("Got", response.StatusCode),
					zap.Int("Expected", http.StatusOK))
				continue
			}
		}
	}
}

func (p *Plane) subscribeToMonitorCommands() {
	defer p.waitGroup.Done()

	// todo: use websockets instead
}

func (p *Plane) SendProcessStatus() error {
	return nil
}

func (p *Plane) WaitUntilCompletion() {
	p.waitGroup.Wait()
}

func (p *Plane) Stop() error {
	p.logger.Debug("Stop control plane")
	p.cancel()
	return nil
}
