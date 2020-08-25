package requests

type RequestDetectSuspectedHangs struct {
	Pid      uint32
	Duration uint64
	Restart  bool
}

func (n *RequestDetectSuspectedHangs) RequestType() int {
	return RequestTypeDetectSuspectedHangs
}
