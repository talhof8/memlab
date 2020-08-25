package requests

const (
	RequestTypeDetectSignals = iota + 1
	RequestTypeDetectThresholds
	RequestTypeDetectSuspectedHangs
)

type DetectionRequest interface {
	RequestType() int
}
