package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/memlab/agent/internal/detection"
	"github.com/memlab/agent/internal/detection/detectors"
	"github.com/memlab/agent/internal/logging"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

var options struct {
	MonitorPid             uint32 `short:"p" long:"monitor-pid" description:"Monitor PID"`
	MaxConcurrentDetectors int    `short:"m" long:"max-detectors" description:"Max concurrent detectors" default:"10"`
	Debug                  bool   `short:"d" long:"debug" description:"Debug mode"`
	ApiUrl                 string `short:"u" long:"api-url" description:"Api URL"`     // todo: move to a config file.
	ApiToken               string `short:"t" long:"api-token" description:"Api token"` // todo: move to a config file.
}

const (
	exitCodeErr = -1
)

var (
	logger              *zap.Logger
	detectionController *detection.Controller
	signalsChan         = make(chan os.Signal)
)

// todo: prettify code

func main() {
	_, err := flags.Parse(&options)
	if err != nil {
		fmt.Printf("Failed to parse arguments: %v\n", err)
		os.Exit(exitCodeErr)
	}

	logger, err = logging.NewLogger("memlab-agent", options.Debug)
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(exitCodeErr)
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
	}()
}

func startAgent() error {
	var err error
	detectionController, err = detection.NewController(logger, options.MaxConcurrentDetectors)
	if err != nil {
		return errors.WithMessage(err, "new misbehavior detection controller")
	}

	err = detectionController.AddDetector(detectors.DetectorTypeSignals, false, options.MonitorPid)
	if err != nil {
		return errors.WithMessagef(err, "adding detector '%s'", detectors.DetectorTypeSignals.Name())
	}

	if err := detectionController.Start(); err != nil {
		return errors.WithMessage(err, "start detection controller")
	}

	detectionController.WaitUntilCompletion()
	return nil
}

func stopAgent() error {
	if detectionController == nil {
		return errors.New("uninitialized misbehavior detection controller")
	}

	if err := detectionController.Stop(); err != nil {
		return errors.WithMessage(err, "stop misbehavior detection controller")
	}

	return nil
}
