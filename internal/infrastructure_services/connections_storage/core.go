package connections_storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	xerrors "github.com/syth0le/gopnik/errors"
	"go.uber.org/zap"

	"github.com/syth0le/realtime-notification-service/internal/clients/redis"
	"github.com/syth0le/realtime-notification-service/internal/model"
)

const (
	UserHashType HashType = "user"
)

type HashType string

func (h HashType) String() string {
	return string(h)
}

type ScanData[T any] struct {
	Data T
}

func (s *ScanData[T]) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *ScanData[T]) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	return nil
}

type Service interface {
	AddConnection(ctx context.Context, userID *model.UserID, conn net.Conn) error
	DeleteConnection(ctx context.Context, userID *model.UserID, conn net.Conn) error
	FlushAllUserConnections(ctx context.Context, userID *model.UserID) error
	GetConnection(ctx context.Context, userID *model.UserID, conn net.Conn) (net.Conn, error)
	GetUserConnections(ctx context.Context, userID *model.UserID) ([]net.Conn, error)
}

type ServiceImpl struct {
	Client redis.Client
	Logger *zap.Logger
}

func (s *ServiceImpl) AddConnection(ctx context.Context, userID *model.UserID, conn net.Conn) error {
	keyHash, err := makeHash(UserHashType, userID.String())
	if err != nil {
		return fmt.Errorf("make hash: %w", err)
	}

	err = s.Client.HSetNX(
		ctx,
		false,
		keyHash,
		conn.RemoteAddr().String(),
		&ScanData[net.Conn]{Data: conn},
	)
	if err != nil {
		return fmt.Errorf("connections_storage set: %w", err)
	}

	s.Logger.Sugar().Infof("key %s saved in connections_storage", keyHash)

	return nil
}

func (s *ServiceImpl) DeleteConnection(ctx context.Context, userID *model.UserID, conn net.Conn) error {
	keyHash, err := makeHash(UserHashType, userID.String())
	if err != nil {
		return fmt.Errorf("make hash: %w", err)
	}

	err = s.Client.HDel(ctx, keyHash, conn.RemoteAddr().String())
	if err != nil {
		return fmt.Errorf("connections_storage delete: %w", err)
	}

	return nil
}

func (s *ServiceImpl) FlushAllUserConnections(ctx context.Context, userID *model.UserID) error {
	keyHash, err := makeHash(UserHashType, userID.String())
	if err != nil {
		return fmt.Errorf("make hash: %w", err)
	}

	err = s.Client.Delete(ctx, keyHash)
	if err != nil {
		return fmt.Errorf("connections_storage delete: %w", err)
	}

	return nil
}

func (s *ServiceImpl) GetConnection(ctx context.Context, userID *model.UserID, conn net.Conn) (net.Conn, error) {
	keyHash, err := makeHash(UserHashType, userID.String())
	if err != nil {
		return nil, fmt.Errorf("make hash: %w", err)
	}

	scanData := &ScanData[net.Conn]{}
	err = s.Client.HGet(ctx, keyHash, conn.RemoteAddr().String(), scanData)
	if err != nil {
		return nil, fmt.Errorf("key %s not found in storage: %w", userID.String(), err)
	}

	return scanData.Data, nil
}

func (s *ServiceImpl) GetUserConnections(ctx context.Context, userID *model.UserID) ([]net.Conn, error) {
	keyHash, err := makeHash(UserHashType, userID.String())
	if err != nil {
		return nil, fmt.Errorf("make hash: %w", err)
	}

	mapConnections, err := s.Client.HGetAll(ctx, keyHash)
	if err != nil {
		return nil, fmt.Errorf("key %s not found in storage: %w", userID.String(), err)
	}

	connections := make([]net.Conn, len(mapConnections))
	acc := 0
	for _, val := range mapConnections {
		scanData := new(ScanData[net.Conn])
		if err = scanData.UnmarshalBinary([]byte(val)); err != nil {
			return nil, fmt.Errorf("unmarshal binary: %w", err)
		}

		connections[acc] = scanData.Data
		acc += 1
	}

	return connections, nil
}

func makeHash(hashType HashType, key string) (string, error) {
	switch hashType {
	case UserHashType:
	default:
		return "", xerrors.WrapInternalError(fmt.Errorf("unexpected hash type: %s", hashType))
	}

	if key == "" {
		return "", xerrors.WrapInternalError(fmt.Errorf("key cannot be empty"))
	}

	return fmt.Sprintf("%s-%s", hashType, key), nil
}
