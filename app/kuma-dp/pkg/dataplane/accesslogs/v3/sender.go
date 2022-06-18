package v3

import (
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
)

const (
	defaultConnectTimeout = 5 * time.Second
)

type sender struct {
	log     logr.Logger
	address string
	conn    net.Conn
}

func (s *sender) Connect() error {
	conn, err := net.DialTimeout("tcp", s.address, defaultConnectTimeout)
	if err != nil {
		return fmt.Errorf("failed to connect to a TCP logging backend: %s: %w", s.address, err)
	}
	s.log.Info("connected to TCP logging backend", "address", s.address)
	s.conn = conn
	return nil
}

func (s *sender) Send(record string) error {
	_, err := s.conn.Write([]byte(record))
	return fmt.Errorf("failed to send a log entry to a TCP logging backend: %s: %w", s.address, err)
}

func (s *sender) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
