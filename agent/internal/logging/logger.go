package logging

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func NewLogger(name string, debugMode bool) (*zap.Logger, error) {
	var (
		logger *zap.Logger
		err error
	)

	if debugMode {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		return nil, errors.WithMessage(err, "new logger")
	}

	return logger.Named(name), nil
}
