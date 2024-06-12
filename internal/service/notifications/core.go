package notifications

import (
	"context"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"

	"github.com/syth0le/realtime-notification-service/internal/clients/rabbit"
	"github.com/syth0le/realtime-notification-service/internal/infrastructure_services/cache"
	"github.com/syth0le/realtime-notification-service/internal/model"
)

type Service interface {
	SubscribeFeedNotifications(ctx context.Context, conn *websocket.Conn, userID model.UserID) error
}

type ServiceImpl struct {
	RedisStorage cache.Service
	Logger       *zap.Logger
	Conn         *rabbitmq.Conn

	Enable       bool
	QueueName    string
	ExchangeName string
}

func (s ServiceImpl) SubscribeFeedNotifications(ctx context.Context, conn *websocket.Conn, userID model.UserID) error {
	s.Logger.Sugar().Infof("handle feed notifications for: %s", userID)

	// /
	publisher, err := rabbitmq.NewPublisher(s.Conn)
	if err != nil {
		return err
	}

	post := &model.Post{
		ID:       model.PostID("234"),
		Text:     "SOME TEXT",
		AuthorID: model.UserID("234"),
	}
	binary, err := post.MarshalBinary()
	if err != nil {
		return err
	}

	err = publisher.Publish(
		binary,
		[]string{userID.String()},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsExchange(s.ExchangeName),
	)
	if err != nil {
		return err
	}
	// ////

	consumer, err := rabbit.NewRabbitConsumer(s.Logger, s.Enable, s.QueueName, s.ExchangeName, userID.String(), s.Conn)
	if err != nil {
		return fmt.Errorf("new rabbit consumer: %w", err)
	}

	// TODO: save connected user to redis
	err = conn.WriteMessage(websocket.TextMessage, []byte("TEXT"))
	if err != nil {
		return nil
	}

	err = consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
		log.Printf("consumed: %v", string(d.Body))

		post := new(model.Post)
		err := post.UnmarshalBinary(d.Body)
		if err != nil {
			return rabbitmq.NackDiscard
		}

		err = conn.WriteMessage(websocket.TextMessage, d.Body)
		if err != nil {
			return rabbitmq.NackDiscard
		}

		// rabbitmq.Ack, rabbitmq.NackDiscard, rabbitmq.NackRequeue
		return rabbitmq.Ack
	})
	if err != nil {
		return fmt.Errorf("consumer run: %w", err)
	}

	return consumer.Close()
}
