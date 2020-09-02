package control

import (
	"context"
	"github.com/memlab/agent/internal/detection"
	"github.com/memlab/agent/internal/detection/requests"
	operatorsPkg "github.com/memlab/agent/internal/operations/operators"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var errFailedToConvertInterface = errors.New("failed to convert interface to request obj")

type DetectionRequestsHandler struct {
	detectionController *detection.Controller
}

func NewDetectionRequestsHandler(detectionController *detection.Controller) *DetectionRequestsHandler {
	return &DetectionRequestsHandler{
		detectionController: detectionController,
	}
}

func (d *DetectionRequestsHandler) Start() error {
	return d.detectionController.Start()
}

func (d *DetectionRequestsHandler) Handle(ctx context.Context, rootLogger *zap.Logger,
	detectionRequest requests.DetectionRequest) error {
	var (
		addDetector        bool
		detectionOperators []operatorsPkg.Operator
	)

	switch detectionRequest.RequestType() {
	case requests.RequestTypeDetectSignals:
		detectSignalsRequest, ok := detectionRequest.(*requests.DetectSignals)
		if !ok {
			return errFailedToConvertInterface
		}

		detectionOperators = []operatorsPkg.Operator{
			&operatorsPkg.CollectMetadata{},
		}
		addDetector = detectSignalsRequest.TurnedOn
	case requests.RequestTypeDetectThresholds, requests.RequestTypeDetectSuspectedHangs:
		return nil // todo: currently it's a stub to avoid errors, replace when implementing those detectors.
	default:
		return errors.Errorf("invalid detector type for request type '%d'", detectionRequest.RequestType())
	}

	if !addDetector {
		return d.detectionController.RemoveDetector(detectionRequest, detectionOperators)
	}
	return d.detectionController.AddDetector(detectionRequest, detectionOperators, true)
}

func (d *DetectionRequestsHandler) Stop() error {
	return d.detectionController.Stop()
}

func (d *DetectionRequestsHandler) WaitUntilCompletion() {
	d.detectionController.WaitUntilCompletion()
}
