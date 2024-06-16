package consumers_pool

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"

	"github.com/syth0le/realtime-notification-service/internal/clients/rabbit"
	"github.com/syth0le/realtime-notification-service/internal/infrastructure_services/connections_pool"
	"github.com/syth0le/realtime-notification-service/internal/model"
)

type Service interface {
	AddConsumer(userID *model.UserID) error
}

type ServiceImpl struct {
	logger *zap.Logger

	pool map[model.UserID]rabbit.Consumer

	mutex sync.Mutex

	conn *rabbitmq.Conn

	enable       bool
	queueName    string
	exchangeName string

	connectionsPoolService connections_pool.Service
}

func NewServiceImpl(
	logger *zap.Logger,
	conn *rabbitmq.Conn,
	enable bool,
	queueName string,
	exchangeName string,
	connectionsPoolService connections_pool.Service,
) *ServiceImpl {
	return &ServiceImpl{
		logger:                 logger,
		pool:                   make(map[model.UserID]rabbit.Consumer),
		mutex:                  sync.Mutex{},
		conn:                   conn,
		enable:                 enable,
		queueName:              queueName,
		exchangeName:           exchangeName,
		connectionsPoolService: connectionsPoolService,
	}
}

func (s *ServiceImpl) AddConsumer(userID *model.UserID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.pool[*userID]; ok {
		return nil
	}

	consumer, err := rabbit.NewRabbitConsumer(
		s.logger,
		s.enable,
		s.queueName,
		s.exchangeName,
		userID.String(),
		s.conn,
	)
	if err != nil {
		return fmt.Errorf("new rabbit consumer for user: %s: %w", userID, err)
	}

	s.pool[*userID] = consumer

	go s.runConsumer(consumer, userID)
	return nil
}

func (s *ServiceImpl) runConsumer(consumer rabbit.Consumer, userID *model.UserID) error {
	defer consumer.Close()

	err := consumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
		log.Printf("consumed: %v", string(d.Body))

		post := new(model.Post)
		err := post.UnmarshalBinary(d.Body)
		if err != nil {
			s.logger.Sugar().Errorf("unmarshal binary: %v", err)
			return rabbitmq.NackDiscard
		}

		connections, err := s.connectionsPoolService.GetUserConnections(userID)
		if err != nil {
			s.logger.Sugar().Errorf("get user connections: %v", err)
			return rabbitmq.NackDiscard
		}

		if len(connections) == 0 {
			s.logger.Sugar().Infof("empty list of connections: %v", err)
			s.pool[*userID].Close()
			delete(s.pool, *userID)
			return rabbitmq.NackRequeue
		}

		for _, conn := range connections {
			err = wsutil.WriteServerMessage(conn, ws.OpText, d.Body)
			if err != nil {
				s.logger.Sugar().Errorf("write message body: %v", err)
				err := s.connectionsPoolService.DeleteConnection(userID, conn)
				if err != nil {
					s.logger.Sugar().Errorf("delete connection: %v", err)
					return rabbitmq.NackDiscard
				}

				return rabbitmq.NackDiscard
			}
		}

		return rabbitmq.Ack
	})

	return err
}

func (s *ServiceImpl) UselessCode(userID *model.UserID) {
	publisher, _ := rabbitmq.NewPublisher(s.conn)

	go func() {
		for {
			res, _ := s.connectionsPoolService.GetUserConnections(userID)

			for _, conn := range res {
				s.logger.Sugar().Infof("user connection: %v %v", conn.RemoteAddr(), conn.LocalAddr())
			}

			defer publisher.Close()

			post := &model.Post{
				ID:       model.PostID("234"),
				Text:     "SOME TEXT",
				AuthorID: model.UserID("234"),
			}
			binary, _ := post.MarshalBinary()

			publisher.Publish(
				binary,
				[]string{userID.String()},
				rabbitmq.WithPublishOptionsContentType("application/json"),
				rabbitmq.WithPublishOptionsExchange(s.exchangeName),
				rabbitmq.WithPublishOptionsExpiration("5000"),
			)

			time.Sleep(5 * time.Second)
		}
	}()
}
