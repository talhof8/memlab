package detection

import (
	"github.com/memlab/agent/internal/detection/detectors"
	"github.com/memlab/agent/internal/detection/requests"
)

var requestTypeToDetectorType = map[requests.RequestType]detectors.DetectorType{
	requests.RequestTypeDetectSignals: detectors.DetectorTypeSignals,
}
