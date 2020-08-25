package requests

type RequestDetectSignals struct {
	Pid      uint32
	TurnedOn bool
	Restart  bool
}

func (n *RequestDetectSignals) RequestType() int {
	return RequestTypeDetectSignals
}
