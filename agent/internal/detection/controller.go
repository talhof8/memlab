package detection

import (
	"context"
	"github.com/memlab/agent/internal/detection/detectors"
	"github.com/memlab/agent/internal/detection/requests"
	"github.com/memlab/agent/internal/operations/operators"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sync"
)

func errDetectorAlreadyExists(detectorName string) error {
	return errors.Errorf("detector '%s' already exists", detectorName)
}

func errDetectorDoesNotExist(detectorName string) error {
	return errors.Errorf("detector '%s' does not exist", detectorName)
}

type Controller struct {
	logger             *zap.Logger
	waitGroup          sync.WaitGroup
	context            context.Context
	cancel             context.CancelFunc
	detectors          map[string]detectors.Detector
	lock               sync.RWMutex
	detectorsSemaphore chan int
	mergedReportsChan  chan map[string]interface{}
}

func NewController(rootLogger *zap.Logger, maxConcurrentDetectors int) (*Controller, error) {
	logger := rootLogger.Named("detection-controller")

	ctx, cancel := context.WithCancel(context.Background())
	return &Controller{
		logger:             logger,
		context:            ctx,
		cancel:             cancel,
		detectors:          make(map[string]detectors.Detector, 0),
		detectorsSemaphore: make(chan int, maxConcurrentDetectors),
		mergedReportsChan:  make(chan map[string]interface{}, 0),
	}, nil
}

func (c *Controller) AddDetector(detectionRequest requests.DetectionRequest, detectionOperators []operators.Operator,
	start bool) error {
	detectorType, err := c.detectorType(detectionRequest)
	if err != nil {
		return err
	}

	detectorName := detectorType.Name()

	c.logger.Debug("Add detector", zap.String("DetectorName", detectorName))

	detector, err := detectors.NewDetector(detectorType, c.context, c.logger, detectionRequest, detectionOperators)
	if err != nil {
		return errors.WithMessage(err, "new detector")
	}

	if _, exists := c.detectors[detectorName]; exists {
		return errDetectorAlreadyExists(detectorName)
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if _, exists := c.detectors[detectorName]; exists { // Double-checked locking.
		return errDetectorAlreadyExists(detectorName)
	}

	c.detectors[detectorName] = detector

	if start {
		c.logger.Debug("Detector is not running, running it since 'start' flag is turned-on",
			zap.String("DetectorName", detectorName))

		c.startDetector(detector)
	}
	return nil
}

func (c *Controller) RemoveDetector(detectionRequest requests.DetectionRequest) error {
	detectorType, err := c.detectorType(detectionRequest)
	if err != nil {
		return err
	}

	detectorName := detectorType.Name()

	c.logger.Debug("Remove detector", zap.String("DetectorName", detectorName))

	if _, exists := c.detectors[detectorName]; !exists {
		return errDetectorDoesNotExist(detectorName)
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	detector, exists := c.detectors[detectorName]
	if !exists { // Double-checked locking.
		return errDetectorDoesNotExist(detectorName)
	}

	if detector.Running() {
		c.logger.Debug("Detector is running, stopping it", zap.String("DetectorName", detectorName))
		if err := detector.StopDetection(); err != nil {
			return errors.WithMessagef(err, "stop detector '%s'", detectorName)
		}
	}

	delete(c.detectors, detectorName)

	return nil
}

func (c *Controller) detectorType(detectionRequest requests.DetectionRequest) (detectors.DetectorType, error) {
	var detectorType detectors.DetectorType

	switch detectionRequest.RequestType() {
	case requests.RequestTypeDetectSignals:
		detectorType = detectors.DetectorTypeSignals
	default:
		return 0, errors.Errorf("invalid detector type for request type '%d'", detectionRequest.RequestType())
	}

	return detectorType, nil
}

func (c *Controller) Start() error {
	c.logger.Debug("Start detection controller")

	c.lock.RLock()
	defer c.lock.RUnlock()

	for _, detector := range c.detectors {
		c.startDetector(detector) // Note: must not block otherwise RLock() will be blocked as well.
	}

	return nil
}

// Will block if reached max capacity.
func (c *Controller) acquireDetectorsSemaphoreBlocking() {
	c.detectorsSemaphore <- 1
}

func (c *Controller) releaseDetectorsSemaphore() {
	<-c.detectorsSemaphore
}

func (c *Controller) startDetector(detector detectors.Detector) {
	c.acquireDetectorsSemaphoreBlocking()
	c.waitGroup.Add(1)

	go func() {
		funcLogger := c.logger.With(zap.String("DetectorName", detector.DetectorName()))
		defer c.waitGroup.Done()
		defer c.releaseDetectorsSemaphore()

		funcLogger.Debug("Start detection loop")
		defer funcLogger.Debug("Done detection loop")

		// Spawn before starting detection to avoid races.
		c.waitGroup.Add(1)
		go c.mergeDetectorReportsChan(detector)

		err := detector.StartDetectionLoop()
		if err != nil {
			funcLogger.Error("Failed to start detection for detector", zap.Error(err))
			return
		}

		detector.WaitUntilCompletion()
	}()
}

func (c *Controller) mergeDetectorReportsChan(detector detectors.Detector) {
	defer c.waitGroup.Done()

	for {
		select {
		case <-c.context.Done():
			close(c.mergedReportsChan)
			return
		case mergedReport, ok := <-detector.MergedReportsChan():
			if !ok {
				return
			}
			c.mergedReportsChan <- mergedReport
		}
	}
}

func (c *Controller) WaitUntilCompletion() {
	c.waitGroup.Wait() // Block until all detectors are done.
}

func (c *Controller) Stop() error {
	c.logger.Debug("Stop detection controller")
	c.cancel() // Will cancel all child-contexts passed to detectors.
	return nil
}

func (c *Controller) MergedReportsChan() <-chan map[string]interface{} {
	return c.mergedReportsChan
}
