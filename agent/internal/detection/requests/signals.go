package requests

import "github.com/memlab/agent/internal/types"

type DetectSignals struct {
	Pid      types.Pid
	TurnedOn bool
	Restart  bool
}

func (n *DetectSignals) RequestType() int {
	return RequestTypeDetectSignals
}
