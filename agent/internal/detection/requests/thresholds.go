package requests

import (
	"fmt"
	"github.com/memlab/agent/internal/types"
)

type DetectThresholds struct {
	Pid                      types.Pid
	CpuThreshold             int
	MemoryThreshold          int
	RestartOnCpuThreshold    bool
	RestartOnMemoryThreshold bool
	TurnedOn                 bool
}

func (n *DetectThresholds) RequestType() RequestType {
	return RequestTypeDetectThresholds
}

func (n *DetectThresholds) Name() string {
	return fmt.Sprintf("%d.%d", n.RequestType(), n.Pid)
}
