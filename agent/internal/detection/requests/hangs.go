package requests

import (
	"fmt"
	"github.com/memlab/agent/internal/types"
)

type DetectSuspectedHangs struct {
	Pid      types.Pid
	Duration uint64
	Restart  bool
	TurnedOn bool
}

func (n *DetectSuspectedHangs) RequestType() RequestType {
	return RequestTypeDetectSuspectedHangs
}

func (n *DetectSuspectedHangs) Name() string {
	return fmt.Sprintf("%d.%d", n.RequestType(), n.Pid)
}
