package control

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	"github.com/memlab/agent/internal/control/client"
	"github.com/memlab/agent/internal/control/client/responses"
	"github.com/memlab/agent/internal/detection"
	"github.com/memlab/agent/internal/reports"
	generalReports "github.com/memlab/agent/internal/reports/general"
	statePkg "github.com/memlab/agent/internal/state"
	"github.com/memlab/agent/internal/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const (
	endpointHosts            = "hosts"
	endpointProcesses        = "processes"
	endpointDetectionConfigs = "detection_configs"
)

type Plane struct {
	logger                   *zap.Logger
	context                  context.Context
	cancel                   context.CancelFunc
	waitGroup                sync.WaitGroup
	client                   *client.RestfulClient
	config                   *PlaneConfig
	state                    *statePkg.State
	detectionRequestsHandler *DetectionRequestsHandler
	machineId                string
}

func NewPlane(rootLogger *zap.Logger, config *PlaneConfig, detectionController *detection.Controller) (*Plane, error) {
	if valid, err := config.Valid(); !valid {
		return nil, errors.WithMessage(err, "validate control plane config")
	}

	logger := rootLogger.Named("control-plane")
	ctx, cancel := context.WithCancel(context.Background())

	restfulClient, err := client.NewRestfulClient(ctx, logger, config.ApiConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "new restful client")
	}

	machineId, err := machineid.ID() // todo: find a fallback on error (should be a constant identifier)
	if err != nil {
		return nil, errors.WithMessage(err, "get machine id")
	}

	state := statePkg.NewState()
	detectionRequestsHandler := NewDetectionRequestsHandler(detectionController)

	return &Plane{
		context:                  ctx,
		cancel:                   cancel,
		client:                   restfulClient,
		config:                   config,
		state:                    state,
		detectionRequestsHandler: detectionRequestsHandler,
		machineId:                machineId,
	}, nil
}

func (p *Plane) Start() error {
	p.logger.Debug("Start control plane")

	p.waitGroup.Add(1)
	go p.startDetectionRequestsHandler()

	p.waitGroup.Add(1)
	go p.handleDetectionRequests()

	p.waitGroup.Add(1)
	go p.fetchDetectionConfigs()

	p.waitGroup.Add(1)
	go p.reportHostStatus()

	p.waitGroup.Add(1)
	go p.reportProcessList()

	return nil
}

func (p *Plane) startDetectionRequestsHandler() {
	defer p.waitGroup.Done()

	if err := p.detectionRequestsHandler.Start(); err != nil {
		p.logger.Error("Failed to start detection requests handler", zap.Error(err))
		p.cancel()
	}
	p.detectionRequestsHandler.WaitUntilCompletion()
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
			report, err := generalReports.NewHostStatusReport(p.machineId)
			if err != nil {
				p.logger.Error("Failed to create host status report", zap.Error(err))
				continue
			}

			if err := p.postReport(endpointHosts, report); err != nil {
				p.logger.Error("Failed to post report", zap.Error(err))
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
			report, err := generalReports.NewProcessListReport()
			if err != nil {
				p.logger.Error("Failed to create process list report", zap.Error(err))
				continue
			}

			if err := p.postReport(endpointProcesses, report); err != nil {
				p.logger.Error("Failed to post report", zap.Error(err))
			}
		}
	}
}

func (p *Plane) postReport(endpoint string, report reports.Report) error {
	data, err := report.DumpReport()
	if err != nil {
		return errors.WithMessage(err, "dump report")
	}

	response, err := p.client.Post(endpoint, data)
	if err != nil {
		return err
	}

	_ = p.validateResponse(response, http.StatusCreated)
	return nil
}

func (p *Plane) fetchDetectionConfigs() {
	defer p.waitGroup.Done()

	// todo: use websockets instead of polling

	ticker := time.NewTicker(p.config.DetectionConfigurationsPollingInterval)
	for {
		select {
		case <-p.context.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			detectionConfigs, success := p.fetchDetectionConfigsFromBackend()
			if !success {
				continue
			}

			p.state.AddDetectionConfigs(detectionConfigs)
		}
	}
}

func (p *Plane) handleDetectionRequests() {
	defer p.waitGroup.Done()

	for {
		select {
		case <-p.context.Done():
			return
		case detectionRequest, ok := <-p.state.DetectionRequests():
			if !ok {
				p.logger.Error("Detection requests channel was closed unexpectedly")
				p.cancel()
				return // todo: re-open instead of returning?
			}

			if err := p.detectionRequestsHandler.Handle(p.context, p.logger, detectionRequest); err != nil {
				p.logger.Error("Failed to handle detection request", zap.Error(err),
					zap.Int("RequestType", detectionRequest.RequestType()))
			}
		}
	}
}

func (p *Plane) fetchDetectionConfigsFromBackend() (map[types.Pid]*responses.DetectionConfiguration, bool) {
	endpoint := fmt.Sprintf("%s/by_machine/%s/", endpointDetectionConfigs, p.machineId)
	bodyBytes, success := p.fetchFromBackend(endpoint)
	if !success {
		return nil, false
	}

	configList := make([]*responses.DetectionConfiguration, 0)
	if err := json.Unmarshal(bodyBytes, &configList); err != nil {
		p.logger.Error("Failed to parse http response body", zap.Error(err))
		return nil, false
	}

	configs := make(map[types.Pid]*responses.DetectionConfiguration, len(configList))
	for _, detectionConfiguration := range configList {
		configs[detectionConfiguration.Pid] = detectionConfiguration
	}

	return configs, true
}

func (p *Plane) fetchFromBackend(endpoint string) ([]byte, bool) {
	httpResponse, err := p.client.Get(endpoint)
	if err != nil {
		p.logger.Error("Failed to fetch data from backend", zap.Error(err))
		return nil, false
	} else if valid := p.validateResponse(httpResponse, http.StatusOK); !valid {
		return nil, false
	}

	defer func() {
		if err := httpResponse.Body.Close(); err != nil {
			p.logger.Error("Failed to close http response body", zap.Error(err))
		}
	}()

	bodyBytes, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		p.logger.Error("Failed to read http response body", zap.Error(err))
		return nil, false
	}
	return bodyBytes, true
}

func (p *Plane) validateResponse(response *http.Response, desiredStatus int) bool {
	if response.StatusCode != desiredStatus {
		p.logger.Warn("Got a bad status code", zap.Int("Got", response.StatusCode),
			zap.Int("Expected", desiredStatus))
		return false
	}
	return true
}

func (p *Plane) ReportProcessEvent() error {
	return nil // todo: implement
}

func (p *Plane) WaitUntilCompletion() {
	p.waitGroup.Wait()
}

func (p *Plane) Stop() error {
	p.logger.Debug("Stop control plane")

	if err := p.detectionRequestsHandler.Stop(); err != nil {
		return errors.WithMessage(err, "stop detection requests handler")
	}

	p.cancel()
	return nil
}
