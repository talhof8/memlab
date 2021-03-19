package requests

import (
	"fmt"
	"github.com/memlab/agent/internal/types"
)

type DetectSignals struct {
	Pid      types.Pid
	Restart  bool
	TurnedOn bool
}

func (n *DetectSignals) RequestType() RequestType {
	return RequestTypeDetectSignals
}

func (n *DetectSignals) Name() string {
	return fmt.Sprintf("%d.%d", n.RequestType(), n.Pid)
}
