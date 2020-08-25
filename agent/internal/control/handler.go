package control

import (
	"github.com/memlab/agent/internal/detection"
	"github.com/memlab/agent/internal/detection/requests"
	"github.com/pkg/errors"
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

func (d *DetectionRequestsHandler) Handle(detectionRequest requests.DetectionRequest) error {
	var isAddOperation bool

	switch detectionRequest.RequestType() {
	case requests.RequestTypeDetectSignals:
		detectSignalsRequest, ok := detectionRequest.(*requests.RequestDetectSignals)
		if !ok {
			return errFailedToConvertInterface
		}

		isAddOperation = detectSignalsRequest.TurnedOn
	default:
		return errors.Errorf("invalid detector type for request type '%d'", detectionRequest.RequestType())
	}

	if isAddOperation {
		return d.detectionController.AddDetector(detectionRequest, true)
	}
	return d.detectionController.RemoveDetector(detectionRequest)
}

func (d *DetectionRequestsHandler) Stop() error {
	return d.detectionController.Stop()
}

func (d *DetectionRequestsHandler) WaitUntilCompletion() {
	d.detectionController.WaitUntilCompletion()
}
