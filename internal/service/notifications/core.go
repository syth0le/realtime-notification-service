package notifications

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"

	"github.com/syth0le/realtime-notification-service/internal/infrastructure_services/connections_pool"
	"github.com/syth0le/realtime-notification-service/internal/model"
)

type Service interface {
	SubscribeFeedNotifications(ctx context.Context, conn net.Conn, userID *model.UserID) error
}

type ServiceImpl struct {
	ConnectionsPool connections_pool.Service
	Logger          *zap.Logger
	Conn            *rabbitmq.Conn

	Enable       bool
	QueueName    string
	ExchangeName string
}

func (s ServiceImpl) SubscribeFeedNotifications(ctx context.Context, conn net.Conn, userID *model.UserID) error {
	s.Logger.Sugar().Infof("handle feed notifications for: %s", userID)

	s.ConnectionsPool.AddConnection(userID, conn)

	defer s.ConnectionsPool.DeleteConnection(userID, conn)

	res, err := s.ConnectionsPool.GetUserConnections(userID)
	if err != nil {
		return fmt.Errorf("get user connections: %w", err)
	}
	for _, conn := range res {
		s.Logger.Sugar().Infof("user connection: %v %v", conn.RemoteAddr(), conn.LocalAddr())
	}

	// {
	// 	consumer, err := rabbit.NewRabbitConsumer(s.Logger, s.Enable, s.QueueName, s.ExchangeName, userID.String(), s.Conn)
	// 	if err != nil {
	// 		return fmt.Errorf("new rabbit consumer: %w", err)
	// 	}
	// 	defer consumer.Close()
	//
	// 	publisher, err := rabbitmq.NewPublisher(s.Conn)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer publisher.Close()
	//
	// 	post := &model.Post{
	// 		ID:       model.PostID("234"),
	// 		Text:     "SOME TEXT",
	// 		AuthorID: model.UserID("234"),
	// 	}
	// 	binary, err := post.MarshalBinary()
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	err = publisher.Publish(
	// 		binary,
	// 		[]string{userID.String()},
	// 		rabbitmq.WithPublishOptionsContentType("application/json"),
	// 		rabbitmq.WithPublishOptionsExchange(s.ExchangeName),
	// 	)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	for {
		err = wsutil.WriteServerMessage(conn, ws.OpText, []byte(time.Now().String()))
		if err != nil {
			return fmt.Errorf("write message body: %v", err)
		}

		time.Sleep(5 * time.Second)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	// err = consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
	//
	// 	select {
	// 	case <-ctx.Done():
	// 		return rabbitmq.NackDiscard
	// 	default:
	// 		log.Printf("consumed: %v", string(d.Body))
	//
	// 		post := new(model.Post)
	// 		err := post.UnmarshalBinary(d.Body)
	// 		if err != nil {
	// 			return rabbitmq.NackDiscard
	// 		}
	//
	// 		s.Logger.Sugar().Infof("conn: %v", conn == nil)
	//
	// 		err = wsutil.WriteServerMessage(conn, ws.OpText, d.Body)
	// 		if err != nil {
	// 			s.Logger.Sugar().Errorf("write message body: %v", err)
	// 			return rabbitmq.NackDiscard
	// 		}
	// 	}
	//
	// 	// rabbitmq.Ack, rabbitmq.NackDiscard, rabbitmq.NackRequeue
	// 	return rabbitmq.Ack
	// })
	// if err != nil {
	// 	return fmt.Errorf("consumer run: %w", err)
	// }
}
