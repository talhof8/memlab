package control

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/memlab/agent/internal/client"
	"github.com/memlab/agent/internal/client/models"
	"github.com/memlab/agent/internal/detection"
	"github.com/memlab/agent/internal/host"
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
	endpointProcessEvents    = "process_events"
	endpointDetectionConfigs = "detection_configs"
	maxBackoffRetries        = 10
)

type Plane struct {
	logger                    *zap.Logger
	context                   context.Context
	cancel                    context.CancelFunc
	waitGroup                 sync.WaitGroup
	client                    *client.RestfulClient
	config                    *PlaneConfig
	state                     *statePkg.State
	detectionRequestsHandler  *DetectionRequestsHandler
	machineId                 string
	initialHostStatusReported chan struct{}
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

	machineId, err := host.MachineId()
	if err != nil {
		return nil, err
	}

	state := statePkg.NewState()
	detectionRequestsHandler := NewDetectionRequestsHandler(detectionController)

	return &Plane{
		logger:                    logger,
		context:                   ctx,
		cancel:                    cancel,
		client:                    restfulClient,
		config:                    config,
		state:                     state,
		detectionRequestsHandler:  detectionRequestsHandler,
		machineId:                 machineId,
		initialHostStatusReported: make(chan struct{}, 1),
	}, nil
}

func (p *Plane) Start() error {
	p.logger.Debug("Start control plane")

	// Note: go routines spawning order is important to avoid races.

	p.waitGroup.Add(1)
	go p.reportProcessEvents()

	p.waitGroup.Add(1)
	go p.startDetectionRequestsHandler()

	p.waitGroup.Add(1)
	go p.handleDetectionRequests()

	p.waitGroup.Add(1)
	go p.fetchDetectionConfigs()

	p.waitGroup.Add(1)
	go p.startHostStatusReporter()

	p.waitGroup.Add(1)
	go p.startProcessListReporter()

	return nil
}

func (p *Plane) reportProcessEvents() {
	defer p.waitGroup.Done()

	for {
		detectionReportsChan := p.detectionRequestsHandler.detectionController.DetectionReportsChan()

		select {
		case <-p.context.Done():
			return
		case report, ok := <-detectionReportsChan:
			if !ok {
				p.logger.Error("Detection reports channel was closed unexpectedly")
				p.cancel()
				return
			}

			data, err := json.Marshal(report)
			if err != nil {
				p.logger.Error("Failed to marshal report", zap.Error(err), zap.Any("Report", report))
				continue
			}

			p.logger.Debug("Reporting process event", zap.Any("Data", report))
			if err := p.post(endpointProcessEvents, data); err != nil {
				p.logger.Error("Failed to post event", zap.Error(err))
			}
		}
	}
}

func (p *Plane) startDetectionRequestsHandler() {
	defer p.waitGroup.Done()

	if err := p.detectionRequestsHandler.Start(); err != nil {
		p.logger.Error("Failed to start detection requests handler", zap.Error(err))
		p.cancel()
	}
	p.detectionRequestsHandler.WaitUntilCompletion()
}

func (p *Plane) startHostStatusReporter() {
	defer p.waitGroup.Done()

	ticker := time.NewTicker(p.config.HostStatusReportInterval)

	p.logger.Debug("Reporting host status (initial)")
	p.reportHostStatus() // Report immediately at first call.
	p.initialHostStatusReported <- struct{}{}

	for {
		select {
		case <-p.context.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			p.logger.Debug("Reporting host status (recurring)")
			p.reportHostStatus()
		}
	}
}

func (p *Plane) reportHostStatus() {
	report, err := generalReports.NewHostStatusReport(p.machineId)
	if err != nil {
		p.logger.Error("Failed to create host status report", zap.Error(err))
		return
	}

	if err := p.sendReport(endpointHosts, report); err != nil {
		p.logger.Error("Failed to post report", zap.Error(err))
	}
}

func (p *Plane) startProcessListReporter() {
	defer p.waitGroup.Done()

	// Should only be reported after host info is sent to the backend.
	<-p.initialHostStatusReported
	p.logger.Debug("Reporting process list (initial)")
	p.reportProcessList()

	ticker := time.NewTicker(p.config.ProcessListReportInterval)
	for {
		select {
		case <-p.context.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			p.logger.Debug("Reporting process list (recurring)")
			p.reportProcessList()
		}
	}
}

func (p *Plane) reportProcessList() {
	report, err := generalReports.NewProcessListReport(p.machineId)
	if err != nil {
		p.logger.Error("Failed to create process list report", zap.Error(err))
	}

	if err := p.sendReport(endpointProcesses, report); err != nil {
		p.logger.Error("Failed to post report", zap.Error(err))
	}
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

			for _, detectionConfig := range detectionConfigs {
				if err := p.state.PutDetectionConfig(detectionConfig); err != nil {
					if err == statePkg.ErrExpiredDetectionConfig {
						p.markDetectionConfigIrrelevant(detectionConfig)
					} else {
						p.logger.Error("Failed to put detection config", zap.Error(err))
					}
					continue
				}
			}
		}
	}
}

func (p *Plane) markDetectionConfigIrrelevant(detectionConfig *models.DetectionConfiguration) {
	endpoint := fmt.Sprintf("%s/mark_irrelevant/%s", endpointDetectionConfigs, detectionConfig.ID)

	if err := p.post(endpoint, nil); err != nil {
		p.logger.Error("Failed to mark detection config as irrelevant", zap.Error(err))
	}
}

func (p *Plane) handleDetectionRequests() {
	defer p.waitGroup.Done()

	for {
		select {
		case <-p.context.Done():
			return
		case detectionRequest, ok := <-p.state.DetectionRequestsChan():
			if !ok {
				p.logger.Error("Detection requests channel was closed unexpectedly")
				p.cancel()
				return
			}

			p.logger.Debug("Got detection request", zap.Int("Type", detectionRequest.RequestType().Int()))
			if err := p.detectionRequestsHandler.Handle(p.context, p.logger, detectionRequest); err != nil {
				p.logger.Error("Failed to handle detection request", zap.Error(err),
					zap.Int("RequestType", detectionRequest.RequestType().Int()))
			}
		}
	}
}

func (p *Plane) fetchDetectionConfigsFromBackend() (map[types.Pid]*models.DetectionConfiguration, bool) {
	endpoint := fmt.Sprintf("%s/by_machine/%s/", endpointDetectionConfigs, p.machineId)
	bodyBytes, success := p.fetchFromBackend(endpoint)
	if !success {
		return nil, false
	}

	configList := make([]*models.DetectionConfiguration, 0)
	if err := json.Unmarshal(bodyBytes, &configList); err != nil {
		p.logger.Error("Failed to parse http response body", zap.Error(err))
		return nil, false
	}

	configs := make(map[types.Pid]*models.DetectionConfiguration, len(configList))
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

func (p *Plane) sendReport(endpoint string, report reports.Report) error {
	data, err := report.DumpReport()
	if err != nil {
		return err
	}

	return p.post(endpoint, data)
}

func (p *Plane) post(endpoint string, data []byte) error {
	backOffPolicy := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), maxBackoffRetries)

	var (
		response *http.Response
		err      error
	)

	err = backoff.Retry(func() error {
		response, err = p.client.Post(endpoint, data)
		if err != nil {
			return err
		}
		return nil
	}, backOffPolicy)

	if err != nil {
		return err
	}

	_ = p.validateResponse(response, http.StatusOK)
	return nil
}

func (p *Plane) validateResponse(response *http.Response, desiredStatus int) bool {
	if response.StatusCode != desiredStatus {
		p.logger.Warn("Got a bad status code", zap.Int("Got", response.StatusCode),
			zap.Int("Expected", desiredStatus))
		return false
	}
	return true
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
