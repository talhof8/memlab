package requests

type RequestType int

const (
	RequestTypeDetectSignals RequestType = iota + 1
	RequestTypeDetectThresholds
	RequestTypeDetectSuspectedHangs
)

func (rt RequestType) Int() int {
	return int(rt)
}

type DetectionRequest interface {
	RequestType() RequestType
}
