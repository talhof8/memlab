package requests

type RequestDetectThresholds struct {
	Pid                      uint32
	CpuThreshold             int
	MemoryThreshold          int
	RestartOnCpuThreshold    bool
	RestartOnMemoryThreshold bool
}

func (n *RequestDetectThresholds) RequestType() int {
	return RequestTypeDetectThresholds
}
