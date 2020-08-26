package control

import (
	"context"
	"github.com/memlab/agent/internal/detection"
	"github.com/memlab/agent/internal/detection/requests"
	"github.com/memlab/agent/internal/operations"
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

		detectionOperators := make()
		addDetector = detectSignalsRequest.TurnedOn
	default:
		return errors.Errorf("invalid detector type for request type '%d'", detectionRequest.RequestType())
	}

	if !addDetector {
		return d.detectionController.RemoveDetector(detectionRequest)
	}

	detectionOperatorsPipeline := operations.NewPipeline(ctx, rootLogger)
	detectionOperatorsPipeline.AddOperators(detectionOperators...)

	return d.detectionController.AddDetector(detectionRequest, true)
}

func (d *DetectionRequestsHandler) Stop() error {
	return d.detectionController.Stop()
}

func (d *DetectionRequestsHandler) WaitUntilCompletion() {
	d.detectionController.WaitUntilCompletion()
}
