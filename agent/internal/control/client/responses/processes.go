package responses

import (
	"github.com/memlab/agent/internal/types"
	"gopkg.in/guregu/null.v3"
)

type DetectionConfiguration struct {
	Pid                      types.Pid `json:"pid"`
	CreatedAt                null.Time `json:"created_at"`
	ModifiedAt               null.Time `json:"modified_at"`
	DetectSignals            bool      `json:"detect_signals"`
	DetectThresholds         bool      `json:"detect_thresholds"`
	DetectSuspectedHangs     bool      `json:"detect_suspected_hangs"`
	CpuThreshold             int       `json:"cpu_threshold"`
	MemoryThreshold          int       `json:"memory_threshold"`
	SuspectedHangDuration    uint64    `json:"suspected_hang_duration"`
	RestartOnSignal          bool      `json:"restart_on_signal"`
	RestartOnCpuThreshold    bool      `json:"restart_on_cpu_threshold"`
	RestartOnMemoryThreshold bool      `json:"restart_on_memory_threshold"`
	RestartOnSuspectedHang   bool      `json:"restart_on_suspected_hang"`
}
