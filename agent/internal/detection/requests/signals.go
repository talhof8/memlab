package requests

import "github.com/memlab/agent/internal/types"

type DetectSignals struct {
	Pid      types.Pid
	Restart  bool
	TurnedOn bool
}

func (n *DetectSignals) RequestType() int {
	return RequestTypeDetectSignals
}
