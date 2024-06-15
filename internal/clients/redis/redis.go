package redis

import (
	"context"
	"encoding"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	xerrors "github.com/syth0le/gopnik/errors"
	"go.uber.org/zap"

	"github.com/syth0le/realtime-notification-service/cmd/realtime/configuration"
)

const (
	defaultClientName = "realtime-notifications"
)

type Client interface {
	HGet(ctx context.Context, key string, field string, scanTo encoding.BinaryUnmarshaler) error
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HSetNX(ctx context.Context, hasTTL bool, key string, field string, value encoding.BinaryUnmarshaler) error
	Delete(ctx context.Context, keys ...string) error
	HDel(ctx context.Context, key string, fields ...string) error
	Close() error
}

type ClientImpl struct {
	Logger             *zap.Logger
	Client             *redis.Client
	ExpirationDuration time.Duration
	MaxListRange       int64
}

func NewRedisClient(logger *zap.Logger, cfg configuration.RedisConfig) Client {
	if !cfg.Enable {
		return &ClientMock{Logger: logger}
	}

	return &ClientImpl{
		Logger: logger,
		Client: redis.NewClient(&redis.Options{
			Addr:       cfg.Address,
			ClientName: defaultClientName,
			Password:   cfg.Password,
			DB:         cfg.Database,
		}),
		ExpirationDuration: cfg.ExpirationDuration,
		MaxListRange:       cfg.MaxListRange,
	}
}

func (c *ClientImpl) Close() error {
	return c.Client.Close()
}

func (c *ClientImpl) HGet(ctx context.Context, key string, field string, scanTo encoding.BinaryUnmarshaler) error {
	resp, err := c.Client.HGet(ctx, key, field).Result()
	if err != nil {
		if err != redis.Nil {
			return xerrors.WrapInternalError(fmt.Errorf("hget error"))
		}

		return xerrors.WrapNotFoundError(err, "not found in connections_storage")
	}

	err = scanTo.UnmarshalBinary([]byte(resp))
	if err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	return nil
}

func (c *ClientImpl) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	resp, err := c.Client.HGetAll(ctx, key).Result()
	if err != nil {
		if err != redis.Nil {
			return nil, xerrors.WrapInternalError(fmt.Errorf("hget error"))
		}

		return nil, xerrors.WrapNotFoundError(err, "not found in connections_storage")
	}

	return resp, nil
}

func (c *ClientImpl) HSetNX(ctx context.Context, hasTTL bool, key string, field string, value encoding.BinaryUnmarshaler) error {
	err := c.Client.HSetNX(ctx, key, field, value).Err()
	if err != nil {
		return err
	}

	if hasTTL {
		c.Client.Expire(ctx, key, c.ExpirationDuration)
	}

	return nil
}

func (c *ClientImpl) Delete(ctx context.Context, keys ...string) error {
	_, err := c.Client.Del(ctx, keys...).Result()
	if err != nil {
		return fmt.Errorf("del keys: %w", err)
	}

	return nil
}

func (c *ClientImpl) HDel(ctx context.Context, key string, fields ...string) error {
	_, err := c.Client.HDel(ctx, key, fields...).Result()
	if err != nil {
		return fmt.Errorf("hdel keys: %w", err)
	}

	return nil
}
