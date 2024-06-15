package rabbit

import (
	"fmt"

	xerrors "github.com/syth0le/gopnik/errors"
	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"
)

type Consumer interface {
	Close() error
	Run(handler rabbitmq.Handler) error
}

type ConsumerImpl struct {
	Conn     *rabbitmq.Conn
	Consumer *rabbitmq.Consumer
}

func NewRabbitConsumer(
	logger *zap.Logger,
	enable bool,
	queueName string,
	exchangeName string,
	routingKey string,
	conn *rabbitmq.Conn,
) (Consumer, error) {
	if !enable {
		return &ConsumerMock{
			Logger: logger,
		}, nil
	}

	consumer, err := rabbitmq.NewConsumer(
		conn,
		queueName,
		rabbitmq.WithConsumerOptionsRoutingKey(routingKey),
		rabbitmq.WithConsumerOptionsExchangeName(exchangeName),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
	)
	if err != nil {
		return nil, xerrors.WrapInternalError(fmt.Errorf("cannot create consumer: %w", err))
	}

	return &ConsumerImpl{
		Conn:     conn,
		Consumer: consumer,
	}, nil
}

func (c *ConsumerImpl) Close() error {
	c.Consumer.Close()
	return nil
}

func (c *ConsumerImpl) Run(handler rabbitmq.Handler) error {
	return c.Consumer.Run(handler)
}
