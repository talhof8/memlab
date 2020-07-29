package main

import (
	"github.com/memlab/agent/internal/errors"
	"github.com/memlab/agent/internal/kernel/communication"
	"github.com/memlab/agent/internal/logging"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

const (
	nlFamily = 25 // todo: get as flag
)

var (
	logger             *zap.Logger
	kernelCommunicator *communication.Communicator // todo: move control over it to internal component
	signalsChan        = make(chan os.Signal)
)

// todo: prettify code

func main() {
	var err error
	logger, err = logging.NewLogger("memlab-agent")
	if err != nil {
		panic(errors.WrappedErrNewLogger(err))
	}

	setupSignalHandling()

	logger.Info("Start agent")
	if err := startAgent(); err != nil {
		logger.Fatal("Failed to start agent", zap.Error(err))
	}
}

func setupSignalHandling() {
	signal.Notify(signalsChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalsChan
		logger.Info("Stop agent")
		if err := stopAgent(); err != nil {
			logger.Fatal("Failed to stop agent", zap.Error(err))
		}

		os.Exit(0)
	}()
}

func startAgent() error {
	var err error
	kernelCommunicator, err = communication.NewCommunicator(nlFamily)
	if err != nil {
		return errors.WrappedErrNewCommunicator(err)
	}

	for caughtSignal := range kernelCommunicator.Signals() {
		logger.Info("Caught signal", zap.Any("Signal", caughtSignal))
	}
	return nil
}

func stopAgent() error {
	if kernelCommunicator == nil {
		return errors.ErrUninitializedCommunicator
	}

	if err := kernelCommunicator.Close(); err != nil {
		return errors.WrappedErrCloseCommunicator(err)
	}

	return nil
}
