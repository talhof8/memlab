package detectors

import (
	"context"
	"github.com/memlab/agent/internal/detection/requests"
	kernelComm "github.com/memlab/agent/internal/kernel/communication"
	"github.com/memlab/agent/internal/operations"
	"github.com/memlab/agent/internal/operations/operators"
	"github.com/memlab/agent/internal/types"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"sync"
)

const (
	nlFamilyNameReceive = "memlab-ktu"
	nlFamilyNameSend    = "memlab-utk"
)

// todo: collect process exit code

type SignalDetector struct {
	detectorType         DetectorType
	logger               *zap.Logger
	context              context.Context
	cancel               context.CancelFunc
	waitGroup            sync.WaitGroup
	detectionOperators   []operators.Operator
	mergedReportsChan    chan map[string]interface{}
	kernelCommunicator   *kernelComm.Communicator
	running              *atomic.Bool
	detectSignalsRequest *requests.DetectSignals
	monitorPid           types.Pid
	monitorPidRaw        uint32
}

func newSignalDetector(detectorType DetectorType, ctx context.Context, rootLogger *zap.Logger,
	detectionRequest requests.DetectionRequest, detectionOperators []operators.Operator) (*SignalDetector, error) {
	detectSignalsRequest, ok := detectionRequest.(*requests.DetectSignals)
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
		detectionOperators:   detectionOperators,
		mergedReportsChan:    make(chan map[string]interface{}),
		kernelCommunicator:   kernelCommunicator,
		running:              atomic.NewBool(false),
		detectSignalsRequest: detectSignalsRequest,
		monitorPid:           detectSignalsRequest.Pid,
		monitorPidRaw:        detectSignalsRequest.Pid.Uint32(),
	}, nil
}

func (sd *SignalDetector) StartDetectionLoop() error {
	sd.waitGroup.Add(1)
	go sd.handleCaughtSignals()

	err := sd.listenForCaughtSignals()
	if err != nil { // Spawns a go-routine internally.
		return err
	}

	sd.startKernelSignalDetection()

	sd.running.Toggle() // Turn on

	return nil
}

func (sd *SignalDetector) listenForCaughtSignals() error {
	if err := sd.kernelCommunicator.ListenForCaughtSignals(); err != nil {
		sd.logger.Error("Failed to listen for caught-signals", zap.Error(err))
		return err
	}
	return nil
}

func (sd *SignalDetector) handleCaughtSignals() {
	defer sd.waitGroup.Done()

	for {
		select {
		case <-sd.context.Done():
			sd.logger.Debug("Done handling caught signals")
			return
		case caughtSignal, ok := <-sd.kernelCommunicator.CaughtSignalsChan():
			if !ok {
				sd.logger.Error("Caught-signals channel was closed unexpectedly")
				return
			}

			sd.logger.Debug("Caught signal", zap.Any("Signal", caughtSignal))
			sd.handleCaughtSignal(caughtSignal)
		}
	}
}

func (sd *SignalDetector) handleCaughtSignal(caughtSignal *kernelComm.PayloadCaughtSignal) {
	defer sd.waitGroup.Done()

	funcLogger := sd.logger.With(zap.Uint32("Pid", caughtSignal.Pid))

	defer func() {
		if err := sd.kernelCommunicator.NotifyHandledSignal(sd.monitorPidRaw); err != nil {
			funcLogger.Error("Failed to notify handled signal", zap.Error(err),
				zap.Any("Signal", caughtSignal))
			return
		}
	}()

	operatorsPipeline := operations.NewPipeline(sd.context, sd.logger, sd.detectionOperators)

	report, err := operatorsPipeline.Start(sd.monitorPid)
	if err != nil {
		funcLogger.Error("Failed to run operators pipeline", zap.Error(err))
		return
	}

	sd.mergedReportsChan <- report
}

func (sd *SignalDetector) startKernelSignalDetection() {
	sd.logger.Debug("Start kernel signal detection for process", zap.Uint32("Pid", sd.monitorPidRaw))
	if err := sd.kernelCommunicator.WatchProcess(sd.monitorPidRaw); err != nil {
		sd.logger.Error("Failed to watch process", zap.Error(err), zap.Uint32("Pid", sd.monitorPidRaw))
		return
	}
}

func (sd *SignalDetector) stopKernelSignalDetection() {
	sd.logger.Debug("Stop kernel signal detection for process", zap.Uint32("Pid", sd.monitorPidRaw))
	if err := sd.kernelCommunicator.UnwatchProcess(sd.monitorPidRaw); err != nil {
		sd.logger.Error("Failed to unwatch process", zap.Error(err), zap.Uint32("Pid", sd.monitorPidRaw))
		return
	}
}

func (sd *SignalDetector) WaitUntilCompletion() {
	sd.waitGroup.Wait() // Block until detection goroutines are done.
	sd.running.Toggle() // Turn off
}

func (sd *SignalDetector) Running() bool {
	return sd.running.Load()
}

func (sd *SignalDetector) StopDetection() error {
	sd.stopKernelSignalDetection()

	sd.cancel()

	if err := sd.kernelCommunicator.Close(); err != nil {
		return errors.WithMessage(err, "close communicator")
	}

	return nil
}

func (sd *SignalDetector) DetectorName() string {
	return sd.detectorType.Name()
}

func (sd *SignalDetector) Operators() []operators.Operator {
	return []operators.Operator{
		&operators.CollectMetadata{},
	}
}

func (sd *SignalDetector) MergedReportsChan() <-chan map[string]interface{} {
	return sd.mergedReportsChan
}
