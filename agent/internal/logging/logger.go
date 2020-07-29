package logging

import "go.uber.org/zap"

func NewLogger(name string) (*zap.Logger, error) {
	logger, err := zap.NewDevelopment() // todo: use production by default, use a debug flag
	if err != nil {
		return nil, err
	}
	logger.Named(name)
	return logger, nil
}
