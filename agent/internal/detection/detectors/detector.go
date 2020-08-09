package detectors

import (
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Detector interface {
	StartDetectionLoop() error
	Running() bool
	StopDetection() error
	Name() string
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

func NewDetector(detectorType DetectorType, ctx context.Context, rootLogger *zap.Logger, monitorPid uint32) (
	Detector, error) {
	switch detectorType {
	case DetectorTypeSignals:
		return newSignalDetector(detectorType, ctx, rootLogger, monitorPid)
	default:
		return nil, errors.Errorf("unknown detector type '%d'", detectorType)
	}
}
