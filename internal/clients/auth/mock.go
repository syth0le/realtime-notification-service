package auth

import (
	"net/http"

	"go.uber.org/zap"
)

type ClientMock struct {
	logger *zap.Logger
}

func NewClientMock(logger *zap.Logger) *ClientMock {
	return &ClientMock{
		logger: logger,
	}
}

func (m *ClientMock) AuthenticationInterceptor(next http.Handler) http.Handler {
	m.logger.Debug("authenticated through mock service")
	return next
}
