package control

import (
	"github.com/memlab/agent/internal/control/client/responses"
	"github.com/memlab/agent/internal/detection/requests"
)

type State struct {
	detectionConfigsCache map[uint32]*responses.DetectionConfiguration
	detectionRequests     chan requests.DetectionRequest
}

func NewState() *State {
	return &State{
		detectionConfigsCache: make(map[uint32]*responses.DetectionConfiguration, 0),
		detectionRequests:     make(chan requests.DetectionRequest, 0),
	}
}

func (s *State) UpdateDetectionConfigsCache(configsFromBackend map[uint32]*responses.DetectionConfiguration) {
	for _, detectionConfig := range configsFromBackend {
		pid := detectionConfig.Pid

		cachedConfig, configured := s.detectionConfigsCache[pid]
		if !configured {
			s.detectionConfigsCache[pid] = detectionConfig
			s.dispatchDetectionRequests(detectionConfig, nil)
			continue
		}

		// Avoid redundant update if config didn't change
		if !detectionConfig.ModifiedAt.After(cachedConfig.ModifiedAt) {
			continue
		}

		cachedConfig.DetectSignals = detectionConfig.DetectSignals
		cachedConfig.DetectThresholds = detectionConfig.DetectThresholds
		cachedConfig.DetectSuspectedHangs = detectionConfig.DetectSuspectedHangs
		cachedConfig.CpuThreshold = detectionConfig.CpuThreshold
		cachedConfig.MemoryThreshold = detectionConfig.MemoryThreshold
		cachedConfig.SuspectedHangDuration = detectionConfig.SuspectedHangDuration
		cachedConfig.RestartOnSignal = detectionConfig.RestartOnSignal

		s.dispatchDetectionRequests(detectionConfig, cachedConfig)
	}
}

func (s *State) dispatchDetectionRequests(newConfig, oldConfig *responses.DetectionConfiguration) {
	if oldConfig == nil {
		if newConfig.DetectSignals {
			s.sendSignalDetectionNotification(newConfig)
		}

		if newConfig.DetectThresholds {
			s.sendThresholdsDetectionNotification(newConfig)
		}

		if newConfig.DetectSuspectedHangs {
			s.sendSuspectedHangsDetectionNotification(newConfig)
		}

		return
	} else {
		if oldConfig.DetectSignals != newConfig.DetectSignals {
			s.sendSignalDetectionNotification(newConfig)
		}

		if oldConfig.DetectThresholds != newConfig.DetectThresholds {
			s.sendSignalDetectionNotification(newConfig)
		}

		if oldConfig.DetectSuspectedHangs != newConfig.DetectSuspectedHangs {
			s.sendSignalDetectionNotification(newConfig)
		}
	}
}

func (s *State) sendSignalDetectionNotification(newConfig *responses.DetectionConfiguration) {
	s.detectionRequests <- &requests.RequestDetectSignals{
		Pid:      newConfig.Pid,
		TurnedOn: true,
		Restart:  newConfig.RestartOnSignal,
	}
}

func (s *State) sendThresholdsDetectionNotification(newConfig *responses.DetectionConfiguration) {
	s.detectionRequests <- &requests.RequestDetectThresholds{
		Pid:                      newConfig.Pid,
		CpuThreshold:             newConfig.CpuThreshold,
		MemoryThreshold:          newConfig.MemoryThreshold,
		RestartOnCpuThreshold:    newConfig.RestartOnCpuThreshold,
		RestartOnMemoryThreshold: newConfig.RestartOnMemoryThreshold,
	}
}

func (s *State) sendSuspectedHangsDetectionNotification(newConfig *responses.DetectionConfiguration) {
	s.detectionRequests <- &requests.RequestDetectSuspectedHangs{
		Pid:      newConfig.Pid,
		Duration: newConfig.SuspectedHangDuration,
		Restart:  newConfig.RestartOnSuspectedHang,
	}
}

func (s *State) DetectionRequests() <-chan requests.DetectionRequest {
	return s.detectionRequests
}
