package state

import (
	"github.com/memlab/agent/internal/client/models"
	"github.com/memlab/agent/internal/detection/requests"
	"github.com/memlab/agent/internal/types"
	"github.com/pkg/errors"
	psUtil "github.com/shirou/gopsutil/process"
)

type State struct {
	detectionConfigsCache map[types.Pid]*models.DetectionConfiguration
	detectionRequestsChan chan requests.DetectionRequest
}

func NewState() *State {
	return &State{
		detectionConfigsCache: make(map[types.Pid]*models.DetectionConfiguration, 0),
		detectionRequestsChan: make(chan requests.DetectionRequest, 0),
	}
}

func (s *State) DetectionRequestsChan() <-chan requests.DetectionRequest {
	return s.detectionRequestsChan
}

func (s *State) PutDetectionConfig(detectionConfig *models.DetectionConfiguration) error {
	if !detectionConfig.IsRelevant {
		return nil
	}

	pid := detectionConfig.Pid

	if err := s.validateDetectionConfig(detectionConfig, pid); err != nil {
		return err
	}

	cachedConfig, configured := s.detectionConfigsCache[pid]
	if !configured {
		s.detectionConfigsCache[pid] = detectionConfig
		s.dispatchDetectionRequests(detectionConfig, nil)
		return nil
	}

	// Avoid redundant update if config didn't change
	if !detectionConfig.ModifiedAt.Time.After(cachedConfig.ModifiedAt.Time) {
		return nil
	}

	s.dispatchDetectionRequests(detectionConfig, cachedConfig)

	// Only update cached config after new one was dispatched
	s.detectionConfigsCache[pid] = detectionConfig
	return nil
}

/// Validates detection config by comparing the given pid in the detection configuration,
/// to the matching process information from the host.
func (s *State) validateDetectionConfig(detectionConfig *models.DetectionConfiguration, pid types.Pid) error {
	process := &models.Process{Pid: pid}
	liveProcess, err := process.LiveProcess()
	if err != nil {
		if err == psUtil.ErrorProcessNotRunning {
			return ErrExpiredDetectionConfig
		}
		return errors.WithMessagef(err, "get live process for pid '%d'", pid)
	}

	createTimeMilliseconds, err := liveProcess.CreateTime()
	if err != nil {
		if err == psUtil.ErrorProcessNotRunning {
			return ErrExpiredDetectionConfig
		}
		return errors.WithMessagef(err, "get create time for pid '%d'", pid)
	}

	createTime := types.JsonTimeFromMillisecondTimestamp(createTimeMilliseconds)

	if !createTime.Equal(detectionConfig.ProcessCreateTime) {
		return ErrExpiredDetectionConfig
	}

	return nil
}

func (s *State) dispatchDetectionRequests(newConfig, oldConfig *models.DetectionConfiguration) {
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

func (s *State) sendSignalDetectionRequest(newConfig *models.DetectionConfiguration) {
	s.detectionRequestsChan <- &requests.DetectSignals{
		Pid:      newConfig.Pid,
		Restart:  newConfig.RestartOnSignal,
		TurnedOn: newConfig.DetectSignals,
	}
}

func (s *State) sendThresholdsDetectionRequest(newConfig *models.DetectionConfiguration) {
	s.detectionRequestsChan <- &requests.DetectThresholds{
		Pid:                      newConfig.Pid,
		CpuThreshold:             newConfig.CpuThreshold,
		MemoryThreshold:          newConfig.MemoryThreshold,
		RestartOnCpuThreshold:    newConfig.RestartOnCpuThreshold,
		RestartOnMemoryThreshold: newConfig.RestartOnMemoryThreshold,
		TurnedOn:                 newConfig.DetectThresholds,
	}
}

func (s *State) sendSuspectedHangsDetectionRequest(newConfig *models.DetectionConfiguration) {
	s.detectionRequestsChan <- &requests.DetectSuspectedHangs{
		Pid:      newConfig.Pid,
		Duration: newConfig.SuspectedHangDuration,
		Restart:  newConfig.RestartOnSuspectedHang,
		TurnedOn: newConfig.DetectSuspectedHangs,
	}
}
