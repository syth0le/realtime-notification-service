package connections_pool

import (
	"fmt"
	"net"
	"sync"

	xerrors "github.com/syth0le/gopnik/errors"
	"go.uber.org/zap"

	"github.com/syth0le/realtime-notification-service/internal/model"
)

type Service interface {
	AddConnection(userID *model.UserID, conn net.Conn)
	DeleteConnection(userID *model.UserID, conn net.Conn) error
	FlushAllUserConnections(userID *model.UserID) error
	FlushAllConnections()
	GetUserConnections(userID *model.UserID) ([]net.Conn, error)
}

type ServiceImpl struct {
	logger *zap.Logger

	pool map[model.UserID]map[net.Addr]net.Conn

	mutex sync.Mutex
}

func NewServiceImpl(logger *zap.Logger) *ServiceImpl {
	return &ServiceImpl{
		logger: logger,
		pool:   make(map[model.UserID]map[net.Addr]net.Conn),
		mutex:  sync.Mutex{},
	}
}

func (s *ServiceImpl) AddConnection(userID *model.UserID, conn net.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.addElem(*userID, conn)
}

func (s *ServiceImpl) DeleteConnection(userID *model.UserID, conn net.Conn) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.delElem(*userID, conn)
}

func (s *ServiceImpl) FlushAllUserConnections(userID *model.UserID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.flushAllConnections(*userID)
}

func (s *ServiceImpl) FlushAllConnections() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for userID, userConns := range s.pool {
		for _, conn := range userConns {
			conn.Close()
		}
		s.pool[userID] = nil
	}
}

func (s *ServiceImpl) GetUserConnections(userID *model.UserID) ([]net.Conn, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.getAllConnections(*userID)
}

func (s *ServiceImpl) addElem(userID model.UserID, conn net.Conn) {
	s.logger.Sugar().Debugf("before add: (%d)  %s", len(s.pool[userID]), userID)

	if _, ok := s.pool[userID]; ok {
		s.pool[userID][conn.RemoteAddr()] = conn
	} else {
		s.pool[userID] = map[net.Addr]net.Conn{conn.RemoteAddr(): conn}
	}

	s.logger.Sugar().Debugf("after add (%d)  %s", len(s.pool[userID]), userID)
}

func (s *ServiceImpl) delElem(userID model.UserID, conn net.Conn) error {
	if _, ok := s.pool[userID]; !ok {
		return xerrors.WrapNotFoundError(fmt.Errorf("not found user id"), "not found user id")
	}

	if _, ok := s.pool[userID][conn.RemoteAddr()]; !ok {
		return xerrors.WrapNotFoundError(fmt.Errorf("not found connection"), "not found connection")
	}

	delete(s.pool[userID], conn.RemoteAddr())
	return nil
}

func (s *ServiceImpl) flushAllConnections(userID model.UserID) error {
	if _, ok := s.pool[userID]; !ok {
		return xerrors.WrapNotFoundError(fmt.Errorf("not found userID"), "not found user id")
	}

	for _, conn := range s.pool[userID] {
		conn.Close()
	}

	delete(s.pool, userID)
	return nil
}

func (s *ServiceImpl) getAllConnections(userID model.UserID) ([]net.Conn, error) {
	if _, ok := s.pool[userID]; !ok {
		return nil, xerrors.WrapNotFoundError(fmt.Errorf("not found userID"), "not found user id")
	}

	res := make([]net.Conn, len(s.pool[userID]))
	acc := 0
	for _, val := range s.pool[userID] {
		res[acc] = val
		acc += 1
	}

	return res, nil
}
