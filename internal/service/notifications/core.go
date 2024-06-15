package notifications

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"

	"github.com/syth0le/realtime-notification-service/internal/infrastructure_services/connections_pool"
	"github.com/syth0le/realtime-notification-service/internal/infrastructure_services/consumers_pool"
	"github.com/syth0le/realtime-notification-service/internal/model"
)

type Service interface {
	SubscribeFeedNotifications(ctx context.Context, conn net.Conn, userID *model.UserID) error
}

type ServiceImpl struct {
	ConnectionsPool connections_pool.Service
	ConsumersPool   consumers_pool.Service
	Logger          *zap.Logger
}

func (s ServiceImpl) SubscribeFeedNotifications(ctx context.Context, conn net.Conn, userID *model.UserID) error {
	s.Logger.Sugar().Infof("handle feed notifications for: %s", userID)

	s.ConnectionsPool.AddConnection(userID, conn)

	err := s.ConsumersPool.AddConsumer(userID)
	if err != nil {
		return fmt.Errorf("add consumer: %w", err)
	}

	s.Logger.Sugar().Infof("REACHED HERE WIHTOUT BLOCKERS")

	return nil
}
