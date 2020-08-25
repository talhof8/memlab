package control

import (
	"github.com/memlab/agent/internal/control/client"
	"github.com/pkg/errors"
	"time"
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
	} else if pc.ProcessListReportInterval <= 0 {
		return false, errors.New("uninitialized process list report interval")
	}

	return true, nil
}
