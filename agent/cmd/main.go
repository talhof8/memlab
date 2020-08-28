package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/memlab/agent/internal/control"
	"github.com/memlab/agent/internal/control/client"
	"github.com/memlab/agent/internal/detection"
	"github.com/memlab/agent/internal/logging"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var options struct {
	MaxConcurrentDetectors int  `short:"m" long:"max-detectors" description:"Max concurrent detectors" default:"5"`
	Debug                  bool `short:"d" long:"debug" description:"Debug mode"`

	// todo: move below to a config file.
	HostStatusReportInterval               time.Duration `short:"hri" long:"host-status-interval" description:"Host status report interval" default:"30s"`
	ProcessListReportInterval              time.Duration `short:"pri" long:"process-list-interval" description:"Process list report interval" default:"5s"`
	DetectionConfigurationsPollingInterval time.Duration `short:"dri" long:"detection-configs-interval" description:"Detection configurations polling interval" default:"5s"`
	ApiUrl                                 string        `short:"u" long:"api-url" description:"Api URL"`
	ApiToken                               string        `short:"t" long:"api-token" description:"Api token"`
}

const (
	exitCodeErr = -1
)

var (
	logger       *zap.Logger
	controlPlane *control.Plane
	signalsChan  = make(chan os.Signal)
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
	detectionController, err := detection.NewController(logger, options.MaxConcurrentDetectors)
	if err != nil {
		return errors.WithMessage(err, "new detection controller")
	}

	apiConfig := &client.ApiConfig{
		Url:   options.ApiUrl,
		Token: options.ApiToken,
	}

	controlPlaneConfig := &control.PlaneConfig{
		ApiConfig:                              apiConfig,
		HostStatusReportInterval:               options.HostStatusReportInterval,
		ProcessListReportInterval:              options.ProcessListReportInterval,
		DetectionConfigurationsPollingInterval: options.DetectionConfigurationsPollingInterval,
	}

	controlPlane, err = control.NewPlane(logger, controlPlaneConfig, detectionController)
	if err != nil {
		return errors.WithMessage(err, "new control plane")
	}

	if err := controlPlane.Start(); err != nil {
		return errors.WithMessage(err, "start control plane")
	}
	controlPlane.WaitUntilCompletion()
	return nil
}

func stopAgent() error {
	if controlPlane == nil {
		return errors.New("uninitialized control plane")
	}

	if err := controlPlane.Stop(); err != nil {
		return errors.WithMessage(err, "stop control plane")
	}

	return nil
}
