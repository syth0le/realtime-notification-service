package rabbit

import (
	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"
)

type ConsumerMock struct {
	Logger *zap.Logger
}

func (m *ConsumerMock) Close() error {
	m.Logger.Debug("closed consumer mock")
	return nil
}

func (m *ConsumerMock) Run(handler rabbitmq.Handler) error {
	m.Logger.Debug("run through rabbitmq mock")
	return nil
}
