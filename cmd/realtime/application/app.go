package application

import (
	"context"
	"fmt"
	"syscall"

	xclients "github.com/syth0le/gopnik/clients"
	xcloser "github.com/syth0le/gopnik/closer"
	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"

	"github.com/syth0le/realtime-notification-service/cmd/realtime/configuration"
	"github.com/syth0le/realtime-notification-service/internal/clients/auth"
	"github.com/syth0le/realtime-notification-service/internal/infrastructure_services/connections_pool"
	"github.com/syth0le/realtime-notification-service/internal/infrastructure_services/consumers_pool"
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
	authClient    auth.Client
	notifications notifications.Service
}

func (a *App) constructEnv(ctx context.Context) (*env, error) {
	conn, err := a.makeRabbitConn(a.Config.Queue) // todo: make conn pool
	if err != nil {
		return nil, fmt.Errorf("make rabbit conn: %w", err)
	}

	connectionsPool := connections_pool.NewServiceImpl(a.Logger)
	consumersPool := consumers_pool.NewServiceImpl(
		a.Logger,
		conn,
		a.Config.Queue.Enable,
		a.Config.Queue.QueueName,
		a.Config.Queue.ExchangeName,
		connectionsPool,
	)

	authClient, err := a.makeAuthClient(ctx, a.Config.AuthClient)
	if err != nil {
		return nil, fmt.Errorf("make auth client: %w", err)
	}

	return &env{
		authClient: authClient,
		notifications: &notifications.ServiceImpl{
			ConnectionsPool: connectionsPool,
			ConsumersPool:   consumersPool,
			Logger:          a.Logger,
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

func (a *App) makeAuthClient(ctx context.Context, cfg configuration.AuthClientConfig) (auth.Client, error) {
	if !cfg.Enable {
		return auth.NewClientMock(a.Logger), nil
	}

	connection, err := xclients.NewGRPCClientConn(ctx, cfg.Conn)
	if err != nil {
		return nil, fmt.Errorf("new grpc conn: %w", err)
	}

	a.Closer.Add(connection.Close)

	return auth.NewAuthImpl(a.Logger, connection), nil
}
