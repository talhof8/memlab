package detectors

import (
	"context"
	"github.com/memlab/agent/internal/detection/requests"
	"github.com/memlab/agent/internal/operations/operators"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Detector interface {
	StartDetectionLoop() error
	StopDetection() error
	WaitUntilCompletion()
	DetectorName() string
	Operators() []operators.Operator
	MergedReportsChan() <-chan map[string]interface{}
}

type DetectorType int

const (
	DetectorTypeSignals DetectorType = iota
)

var (
	detectorNames = map[DetectorType]string{
		DetectorTypeSignals: "signal-detector",
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
	detectionRequest requests.DetectionRequest, detectionOperators []operators.Operator) (Detector, error) {
	switch detectorType {
	case DetectorTypeSignals:
		return newSignalDetector(detectorType, ctx, rootLogger, detectionRequest, detectionOperators)
	default:
		return nil, errors.Errorf("unknown detector type '%d'", detectorType)
	}
}
