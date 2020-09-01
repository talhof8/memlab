package control

import (
	"github.com/memlab/agent/internal/control/client"
	"github.com/pkg/errors"
	"time"
)

const (
	minHostStatusReportInterval               = time.Minute
	minProcessListReportInterval              = time.Second * 30
	minDetectionConfigurationsPollingInterval = time.Second * 5
)

type PlaneConfig struct {
	ApiConfig                              *client.ApiConfig
	HostStatusReportInterval               time.Duration
	ProcessListReportInterval              time.Duration
	DetectionConfigurationsPollingInterval time.Duration
}

func (pc *PlaneConfig) Valid() (bool, error) {
	if pc.HostStatusReportInterval <= 0 {
		return false, errors.New("uninitialized host status report interval")
	} else if pc.HostStatusReportInterval < minHostStatusReportInterval {
		return false, errors.Errorf("below minimum allowed host status report interval (min: '%s')",
			minHostStatusReportInterval.String())
	}

	if pc.ProcessListReportInterval <= 0 {
		return false, errors.New("uninitialized process list report interval")
	} else if pc.ProcessListReportInterval < minProcessListReportInterval {
		return false, errors.Errorf("below minimum allowed process list report interval (min: '%s')",
			minProcessListReportInterval.String())
	}

	if pc.DetectionConfigurationsPollingInterval <= 0 {
		return false, errors.New("uninitialized detection configs polling interval")
	} else if pc.DetectionConfigurationsPollingInterval < minDetectionConfigurationsPollingInterval {
		return false, errors.Errorf("below minimum allowed detection configs polling interval (min: '%s')",
			minDetectionConfigurationsPollingInterval.String())
	}

	return true, nil
}
