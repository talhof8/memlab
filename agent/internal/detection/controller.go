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

type Controller struct {
	logger               *zap.Logger
	waitGroup            sync.WaitGroup
	context              context.Context
	cancel               context.CancelFunc
	requestDetectors     map[string]detectors.Detector
	lock                 sync.RWMutex
	detectorsSemaphore   chan int
	detectionReportsChan chan map[string]interface{}
}

func NewController(rootLogger *zap.Logger, maxConcurrentDetectors int) (*Controller, error) {
	logger := rootLogger.Named("detection-controller")

	ctx, cancel := context.WithCancel(context.Background())
	return &Controller{
		logger:               logger,
		context:              ctx,
		cancel:               cancel,
		requestDetectors:     make(map[string]detectors.Detector, 0),
		detectorsSemaphore:   make(chan int, maxConcurrentDetectors),
		detectionReportsChan: make(chan map[string]interface{}, 0),
	}, nil
}

func (c *Controller) AddDetector(request requests.DetectionRequest, operators []operators.Operator, start bool) error {
	detectorType, err := c.detectorType(request)
	if err != nil {
		return err
	}

	detectorName := detectorType.Name()
	requestName := request.Name()

	funcLogger := c.logger.With(zap.String("RequestName", requestName), zap.String("DetectorName", detectorName))
	funcLogger.Debug("Add detector")

	detector, err := c.newDetector(request, operators, detectorType)
	if err != nil {
		return err
	}

	if _, exists := c.requestDetectors[requestName]; exists {
		return errDetectorAlreadyExists(requestName)
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if _, exists := c.requestDetectors[requestName]; exists { // Double-checked locking.
		return errDetectorAlreadyExists(requestName)
	}

	c.requestDetectors[requestName] = detector

	if start {
		funcLogger.Debug("Starting detector")
		c.startDetector(detector)
	}
	return nil
}

func (c *Controller) RemoveDetector(request requests.DetectionRequest, operators []operators.Operator) error {
	detectorType, err := c.detectorType(request)
	if err != nil {
		return err
	}

	detectorName := detectorType.Name()
	requestName := request.Name()

	funcLogger := c.logger.With(zap.String("RequestName", requestName), zap.String("DetectorName", detectorName))
	funcLogger.Debug("Remove detector")

	c.lock.Lock()
	defer c.lock.Unlock()

	detector, exists := c.requestDetectors[requestName]

	// In case agent was reloaded after detector was started in previous agent run, but now config has changed
	// to stop it, create a new detector object.
	if !exists {
		detector, err = c.newDetector(request, operators, detectorType)
		if err != nil {
			return err
		}
	}

	funcLogger.Debug("Stopping detector")
	if err := detector.StopDetection(); err != nil {
		return errors.WithMessagef(err, "stop detector '%s'", detectorName)
	}

	delete(c.requestDetectors, detectorName)

	return nil
}

func (c *Controller) detectorType(detectionRequest requests.DetectionRequest) (detectors.DetectorType, error) {
	requestType := detectionRequest.RequestType()

	detectorType, found := requestTypeToDetectorType[requestType]
	if !found {
		return 0, errors.Errorf("invalid detector type for request type '%d'", requestType)
	}

	return detectorType, nil
}

func (c *Controller) newDetector(detectionRequest requests.DetectionRequest, detectionOperators []operators.Operator,
	detectorType detectors.DetectorType) (detectors.Detector, error) {
	detector, err := detectors.NewDetector(detectorType, c.context, c.logger, detectionRequest, detectionOperators)
	if err != nil {
		return nil, errors.WithMessage(err, "new detector")
	}
	return detector, nil
}

func (c *Controller) Start() error {
	c.logger.Debug("Start detection controller")

	c.lock.RLock()
	defer c.lock.RUnlock()

	for _, detector := range c.requestDetectors {
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
			return
		case detectionReport, ok := <-detector.ReportsChan():
			if !ok {
				return
			}
			c.detectionReportsChan <- detectionReport
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

func (c *Controller) DetectionReportsChan() <-chan map[string]interface{} {
	return c.detectionReportsChan
}
