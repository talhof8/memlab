package requests

import "github.com/memlab/agent/internal/types"

type DetectSuspectedHangs struct {
	Pid      types.Pid
	Duration uint64
	Restart  bool
	TurnedOn bool
}

func (n *DetectSuspectedHangs) RequestType() int {
	return RequestTypeDetectSuspectedHangs
}
