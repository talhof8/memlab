package requests

import "github.com/memlab/agent/internal/types"

type DetectThresholds struct {
	Pid                      types.Pid
	CpuThreshold             int
	MemoryThreshold          int
	RestartOnCpuThreshold    bool
	RestartOnMemoryThreshold bool
	TurnedOn                 bool
}

func (n *DetectThresholds) RequestType() int {
	return RequestTypeDetectThresholds
}
