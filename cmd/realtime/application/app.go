package application

import (
	"context"
	"fmt"
	"syscall"

	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"

	xcloser "github.com/syth0le/gopnik/closer"

	"github.com/syth0le/realtime-notification-service/cmd/realtime/configuration"
	"github.com/syth0le/realtime-notification-service/internal/clients/redis"
	"github.com/syth0le/realtime-notification-service/internal/service/notifications"
)

type App struct {
	Config *configuration.Config
	Logger *zap.Logger
	Closer *xcloser.Closer
}

func New(cfg *configuration.Config, logger *zap.Logger) *App {
	return &App{
		Config: cfg,
		Logger: logger,
		Closer: xcloser.NewCloser(logger, cfg.Application.GracefulShutdownTimeout, cfg.Application.ForceShutdownTimeout, syscall.SIGINT, syscall.SIGTERM),
	}
}

func (a *App) Run() error {
	ctx, cancelFunction := context.WithCancel(context.Background())
	a.Closer.Add(func() error {
		cancelFunction()
		return nil
	})

	envStruct, err := a.constructEnv(ctx)
	if err != nil {
		return fmt.Errorf("construct env: %w", err)
	}

	httpServer := a.newHTTPServer(envStruct)
	a.Closer.Add(httpServer.GracefulStop()...)

	a.Closer.Run(httpServer.Run()...)
	a.Closer.Wait()
	return nil
}

type env struct {
	notifications notifications.Service
}

func (a *App) constructEnv(ctx context.Context) (*env, error) {
	redisClient := redis.NewRedisClient(a.Logger, a.Config.ConnectionsStorage) // TODO: move to gopnik
	a.Closer.Add(redisClient.Close)

	conn, err := a.makeRabbitConn(a.Config.Queue) // todo: make conn pool
	if err != nil {
		return nil, fmt.Errorf("make rabbit conn: %w", err)
	}

	return &env{
		notifications: &notifications.ServiceImpl{
			RedisStorage: nil,
			Logger:       a.Logger,
			Conn:         conn,
			Enable:       a.Config.Queue.Enable,
			QueueName:    a.Config.Queue.QueueName,
			ExchangeName: a.Config.Queue.ExchangeName,
		},
	}, nil
}

func (a *App) makeRabbitConn(cfg configuration.RabbitConfig) (*rabbitmq.Conn, error) {
	if !cfg.Enable {
		return nil, nil
	}

	conn, err := rabbitmq.NewConn(
		a.Config.Queue.Address,
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create connection: %w", err)
	}

	return conn, nil
}
