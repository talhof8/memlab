package state

import (
	"github.com/memlab/agent/internal/control/client/responses"
	"github.com/memlab/agent/internal/detection/requests"
	"github.com/memlab/agent/internal/types"
)

type State struct {
	detectionConfigsCache map[types.Pid]*responses.DetectionConfiguration
	detectionRequestsChan chan requests.DetectionRequest
}

func NewState() *State {
	return &State{
		detectionConfigsCache: make(map[types.Pid]*responses.DetectionConfiguration, 0),
		detectionRequestsChan: make(chan requests.DetectionRequest, 0),
	}
}

func (s *State) DetectionRequestsChan() <-chan requests.DetectionRequest {
	return s.detectionRequestsChan
}

func (s *State) AddDetectionConfigs(configsFromBackend map[types.Pid]*responses.DetectionConfiguration) {
	for _, detectionConfig := range configsFromBackend {
		pid := detectionConfig.Pid

		cachedConfig, configured := s.detectionConfigsCache[pid]
		if !configured {
			s.detectionConfigsCache[pid] = detectionConfig
			s.dispatchDetectionRequests(detectionConfig, nil)
			continue
		}

		// Avoid redundant update if config didn't change
		if !detectionConfig.ModifiedAt.Time.After(cachedConfig.ModifiedAt.Time) {
			continue
		}

		s.dispatchDetectionRequests(detectionConfig, cachedConfig)

		// Only update cached config after new one was dispatched
		s.detectionConfigsCache[pid] = detectionConfig
	}
}

func (s *State) dispatchDetectionRequests(newConfig, oldConfig *responses.DetectionConfiguration) {
	if oldConfig == nil { // Build initial detection configuration if it's not cached.
		s.sendSignalDetectionRequest(newConfig)
		s.sendThresholdsDetectionRequest(newConfig)
		s.sendSuspectedHangsDetectionRequest(newConfig)
		return
	}

	if oldConfig.DetectSignals != newConfig.DetectSignals {
		s.sendSignalDetectionRequest(newConfig)
	}

	if oldConfig.DetectThresholds != newConfig.DetectThresholds {
		s.sendThresholdsDetectionRequest(newConfig)
	}

	if oldConfig.DetectSuspectedHangs != newConfig.DetectSuspectedHangs {
		s.sendSuspectedHangsDetectionRequest(newConfig)
	}
}

func (s *State) sendSignalDetectionRequest(newConfig *responses.DetectionConfiguration) {
	s.detectionRequestsChan <- &requests.DetectSignals{
		Pid:      newConfig.Pid,
		Restart:  newConfig.RestartOnSignal,
		TurnedOn: newConfig.DetectSignals,
	}
}

func (s *State) sendThresholdsDetectionRequest(newConfig *responses.DetectionConfiguration) {
	s.detectionRequestsChan <- &requests.DetectThresholds{
		Pid:                      newConfig.Pid,
		CpuThreshold:             newConfig.CpuThreshold,
		MemoryThreshold:          newConfig.MemoryThreshold,
		RestartOnCpuThreshold:    newConfig.RestartOnCpuThreshold,
		RestartOnMemoryThreshold: newConfig.RestartOnMemoryThreshold,
		TurnedOn:                 newConfig.DetectThresholds,
	}
}

func (s *State) sendSuspectedHangsDetectionRequest(newConfig *responses.DetectionConfiguration) {
	s.detectionRequestsChan <- &requests.DetectSuspectedHangs{
		Pid:      newConfig.Pid,
		Duration: newConfig.SuspectedHangDuration,
		Restart:  newConfig.RestartOnSuspectedHang,
		TurnedOn: newConfig.DetectSuspectedHangs,
	}
}
