package redis

import (
	"context"
	"encoding"
	"fmt"

	xerrors "github.com/syth0le/gopnik/errors"
	"go.uber.org/zap"
)

type ClientMock struct {
	Logger *zap.Logger
}

func (c *ClientMock) HGet(ctx context.Context, key string, field string, scanTo encoding.BinaryUnmarshaler) error {
	c.Logger.Debug("hget through connections_storage mock")
	return xerrors.WrapNotFoundError(fmt.Errorf("cannot find smth in connections_storage mock"), "not found in connections_storage")

}

func (c *ClientMock) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	c.Logger.Debug("hget through connections_storage mock")
	return nil, xerrors.WrapNotFoundError(fmt.Errorf("cannot find smth in connections_storage mock"), "not found in connections_storage")
}

func (c *ClientMock) HSetNX(ctx context.Context, hasTTL bool, key string, field string, value encoding.BinaryUnmarshaler) error {
	c.Logger.Debug("hsetnx through connections_storage mock")
	return nil
}

func (c *ClientMock) LPush(ctx context.Context, key string, value encoding.BinaryMarshaler) error {
	c.Logger.Debug("lpush through connections_storage mock")
	return nil
}

func (c *ClientMock) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	c.Logger.Debug("lrange through connections_storage mock")
	return nil, xerrors.WrapNotFoundError(fmt.Errorf("cannot find smth in connections_storage mock"), "not found in connections_storage")
}

func (c *ClientMock) LRem(ctx context.Context, key string, value encoding.BinaryMarshaler) error {
	c.Logger.Debug("lrem through connections_storage mock")
	return nil
}

func (c *ClientMock) Delete(ctx context.Context, keys ...string) error {
	c.Logger.Debug("del through connections_storage mock")
	return nil
}

func (c *ClientMock) HDel(ctx context.Context, key string, fields ...string) error {
	c.Logger.Debug("hdel through connections_storage mock")
	return nil
}

func (c *ClientMock) Close() error {
	return nil
}
