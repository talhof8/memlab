package detectors

import (
	"context"
	"github.com/memlab/agent/internal/detection/requests"
	kernelComm "github.com/memlab/agent/internal/kernel/communication"
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
	ReportsChan() <-chan map[string]interface{}
}

var kernelCommunicator *kernelComm.Communicator

func NewDetector(detectorType DetectorType, ctx context.Context, rootLogger *zap.Logger,
	detectionRequest requests.DetectionRequest, detectionOperators []operators.Operator) (Detector, error) {
	switch detectorType {
	case DetectorTypeSignals:
		if kernelCommunicator == nil {
			var err error
			kernelCommunicator, err = kernelComm.NewCommunicator(rootLogger, NlFamilyNameReceive, NlFamilyNameSend)
			if err != nil {
				return nil, errors.WithMessage(err, "new kernel communicator")
			}

			// todo: need to close it when detection controller stops.
		}

		return newSignalDetector(detectorType, ctx, rootLogger, detectionRequest, detectionOperators, kernelCommunicator)
	default:
		return nil, errors.Errorf("unknown detector type '%d'", detectorType)
	}
}
