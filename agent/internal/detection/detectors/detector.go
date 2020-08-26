package detectors

import (
	"context"
	"github.com/memlab/agent/internal/detection/requests"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Detector interface {
	StartDetectionLoop() error
	WaitUntilCompletion()
	Running() bool
	StopDetection() error
	DetectorName() string
}

type DetectorType int

const (
	DetectorTypeSignals DetectorType = iota
)

var (
	detectorNames = map[DetectorType]string{
		DetectorTypeSignals: "Signal Detector",
	}
)

func (dt DetectorType) Name() string {
	name, found := detectorNames[dt]
	if !found {
		return ""
	}
	return name
}

func NewDetector(detectorType DetectorType, ctx context.Context, rootLogger *zap.Logger,
	detectionRequest requests.DetectionRequest) (Detector, error) {
	switch detectorType {
	case DetectorTypeSignals:
		return newSignalDetector(detectorType, ctx, rootLogger, detectionRequest)
	default:
		return nil, errors.Errorf("unknown detector type '%d'", detectorType)
	}
}
