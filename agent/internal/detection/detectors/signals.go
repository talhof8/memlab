package detectors

import (
	"context"
	"fmt"
	"github.com/memlab/agent/internal/detection/requests"
	kernelComm "github.com/memlab/agent/internal/kernel/communication"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"os/exec"
	"sync"
)

const (
	nlFamilyNameReceive = "memlab-ktu"
	nlFamilyNameSend    = "memlab-utk"
)

type SignalDetector struct {
	detectorType         DetectorType
	waitGroup            sync.WaitGroup
	context              context.Context
	cancel               context.CancelFunc
	logger               *zap.Logger
	kernelCommunicator   *kernelComm.Communicator
	running              *atomic.Bool
	detectSignalsRequest *requests.RequestDetectSignals
	monitorPid           uint32
}

func newSignalDetector(detectorType DetectorType, ctx context.Context, rootLogger *zap.Logger,
	detectionRequest requests.DetectionRequest) (*SignalDetector, error) {
	detectSignalsRequest, ok := detectionRequest.(*requests.RequestDetectSignals)
	if !ok {
		return nil, errors.New("failed to convert interface to detection request object")
	}

	logger := rootLogger.Named("signal-detector")

	kernelCommunicator, err := kernelComm.NewCommunicator(logger, nlFamilyNameReceive, nlFamilyNameSend)
	if err != nil {
		return nil, errors.WithMessage(err, "new kernel communicator")
	}

	ctx, cancel := context.WithCancel(ctx)

	return &SignalDetector{
		detectorType:         detectorType,
		logger:               logger,
		context:              ctx,
		cancel:               cancel,
		kernelCommunicator:   kernelCommunicator,
		running:              atomic.NewBool(false),
		detectSignalsRequest: detectSignalsRequest,
		monitorPid:           detectSignalsRequest.Pid,
	}, nil
}

func (kd *SignalDetector) StartDetectionLoop() error {
	kd.waitGroup.Add(1)
	go kd.handleCaughtSignals()

	err := kd.listenForCaughtSignals()
	if err != nil { // Spawns a go-routine internally.
		return err
	}

	kd.startKernelSignalDetection()

	kd.running.Toggle() // Turn on

	return nil
}

func (kd *SignalDetector) listenForCaughtSignals() error {
	if err := kd.kernelCommunicator.ListenForCaughtSignals(); err != nil {
		kd.logger.Error("Failed to listen for caught-signals", zap.Error(err))
		return err
	}
	return nil
}

func (kd *SignalDetector) handleCaughtSignals() {
	defer kd.waitGroup.Done()

	for {
		select {
		case <-kd.context.Done():
			kd.logger.Debug("Done handling caught signals")
			return
		case caughtSignal, ok := <-kd.kernelCommunicator.CaughtSignals():
			if !ok {
				kd.logger.Error("Caught-signals channel was closed unexpectedly")
				return
			}

			kd.logger.Debug("Caught signal", zap.Any("Signal", caughtSignal))
			kd.handleCaughtSignal(caughtSignal)
		}
	}
}

func (kd *SignalDetector) handleCaughtSignal(caughtSignal *kernelComm.PayloadCaughtSignal) {
	// todo: create operators pipeline and move handling outside of detector scope.
	funcLogger := kd.logger.With(zap.Uint32("Pid", caughtSignal.Pid))

	ps, err := process.NewProcess(int32(caughtSignal.Pid))
	if err != nil {
		if errors.Cause(err) == process.ErrorProcessNotRunning {
			funcLogger.Error("Process is not running")
			return
		}

		funcLogger.Error("Failed to create process object", zap.Error(err))
		return
	}

	executablePath, err := ps.Exe()
	if err != nil {
		funcLogger.Error("Failed to get process' executable", zap.Error(err))
		return
	}

	cmdline, err := ps.Cmdline()
	if err != nil {
		funcLogger.Error("Failed to get process' cmdline", zap.Error(err))
		return
	}

	cpuPercent, err := ps.CPUPercent()
	if err != nil {
		funcLogger.Error("Failed to get process' CPU percent", zap.Error(err))
		return
	}

	memPercent, err := ps.MemoryPercent()
	if err != nil {
		funcLogger.Error("Failed to get process' memory percent", zap.Error(err))
		return
	}

	createTime, err := ps.CreateTime()
	if err != nil {
		funcLogger.Error("Failed to get process' create time", zap.Error(err))
		return
	}

	cwd, err := ps.Cwd()
	if err != nil {
		funcLogger.Error("Failed to get process' cwd", zap.Error(err))
		return
	}

	// todo: get create time, cwd, foreground, username, uids, times, ppid, groups, pagefaults, nice, num fds, rlimit,
	// todo: parents, threads count, tgid, open files

	funcLogger.Info("Dummy dump", zap.String("Executable", executablePath), zap.String("Cmdline", cmdline),
		zap.Float64("CPUPercent", cpuPercent), zap.Float32("MemoryPercent", memPercent),
		zap.Int64("CreateTime", createTime), zap.String("Cwd", cwd))

	cmd := exec.Command("sudo", "procdump", "-p", fmt.Sprintf("%d", caughtSignal.Pid))
	output, err := cmd.Output()
	if err != nil {
		funcLogger.Error("Failed to run command", zap.Error(err))
		return
	}

	funcLogger.Debug("Procdump ran successfully", zap.ByteString("Output", output))

	// todo: call on failure as well?
	if err := kd.kernelCommunicator.NotifyHandledSignal(caughtSignal.Pid); err != nil {
		funcLogger.Error("Failed to notify handled signal", zap.Error(err), zap.Any("Signal", caughtSignal))
		return
	}
}

func (kd *SignalDetector) startKernelSignalDetection() {
	kd.logger.Debug("Start kernel signal detection for process", zap.Uint32("Pid", kd.monitorPid))
	if err := kd.kernelCommunicator.WatchProcess(kd.monitorPid); err != nil {
		kd.logger.Error("Failed to watch process", zap.Error(err), zap.Uint32("Pid", kd.monitorPid))
		return
	}
}

func (kd *SignalDetector) stopKernelSignalDetection() {
	kd.logger.Debug("Stop kernel signal detection for process", zap.Uint32("Pid", kd.monitorPid))
	if err := kd.kernelCommunicator.UnwatchProcess(kd.monitorPid); err != nil {
		kd.logger.Error("Failed to unwatch process", zap.Error(err), zap.Uint32("Pid", kd.monitorPid))
		return
	}
}

func (kd *SignalDetector) WaitUntilCompletion() {
	kd.waitGroup.Wait() // Block until detection goroutines are done.
	kd.running.Toggle() // Turn off
}

func (kd *SignalDetector) Running() bool {
	return kd.running.Load()
}

func (kd *SignalDetector) StopDetection() error {
	kd.stopKernelSignalDetection()

	kd.cancel()

	if err := kd.kernelCommunicator.Close(); err != nil {
		return errors.WithMessage(err, "close communicator")
	}

	return nil
}

func (kd *SignalDetector) Name() string {
	return kd.detectorType.Name()
}
