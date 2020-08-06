package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/memlab/agent/internal/kernel/communication"
	"github.com/memlab/agent/internal/logging"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var opts struct {
	MonitorPid uint32 `short:"p" description:"Monitor PID"`
	Debug      bool   `short:"d" long:"debug" description:"Debug mode"`
}

const (
	exitCodeErr         = -1
	nlFamilyNameReceive = "memlab-ktu"
	nlFamilyNameSend    = "memlab-utk"
)

var (
	logger             *zap.Logger
	kernelCommunicator *communication.Communicator // todo: move control over it to internal component
	monitorPid         uint32                      // todo: get monitor pid from some control component (http server for instance) and pass kernel communicator to it
	signalsChan        = make(chan os.Signal)
)

// todo: prettify code

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		fmt.Printf("Failed to parse arguments: %v\n", err)
		os.Exit(exitCodeErr)
	}

	logger, err = logging.NewLogger("memlab-agent", opts.Debug)
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(exitCodeErr)
	}

	monitorPid = opts.MonitorPid

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
	kernelCommunicator, err = communication.NewCommunicator(nlFamilyNameReceive, nlFamilyNameSend, logger)
	if err != nil {
		return errors.WithMessage(err, "new communicator")
	}

	if err := kernelCommunicator.ListenForCaughtSignals(); err != nil {
		logger.Error("Failed to listen for caught-signals from kernel module", zap.Error(err))
		return err
	}

	// todo: remove
	go func() {
		logger.Info("Sleeping for 5 seconds before sending pid", zap.Uint32("PID", monitorPid))
		time.Sleep(time.Second * 5)

		if err := kernelCommunicator.WatchProcess(monitorPid); err != nil {
			logger.Error("Failed to watch process", zap.Error(err), zap.Uint32("PID", monitorPid))
			return
		}

		// todo: test unwatch functionality
	}()

	for caughtSignal := range kernelCommunicator.CaughtSignals() {
		logger.Info("Caught signal", zap.Any("Signal", caughtSignal))

		if err := kernelCommunicator.NotifyHandledSignal(caughtSignal.Pid); err != nil {
			logger.Error("Failed to notify handled signal", zap.Error(err), zap.Any("Signal", caughtSignal))
			continue
		}
	}
	return nil
}

func stopAgent() error {
	if kernelCommunicator == nil {
		return errors.New("uninitialized communicator")
	}

	if err := kernelCommunicator.Close(); err != nil {
		return errors.WithMessage(err, "close communicator")
	}

	return nil
}
