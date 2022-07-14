package v3

import (
	"net"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
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
		return errors.Wrapf(err, "failed to connect to a TCP logging backend: %s", s.address)
	}
	s.log.Info("connected to TCP logging backend", "address", s.address)
	s.conn = conn
	return nil
}

func (s *sender) Send(record []byte) error {
	_, err := s.conn.Write(record)
	return errors.Wrapf(err, "failed to send a log entry to a TCP logging backend: %s", s.address)
}

func (s *sender) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
