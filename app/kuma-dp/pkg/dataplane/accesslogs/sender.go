package accesslogs

import (
	"net"
	"time"

	"github.com/pkg/errors"

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
		return errors.Wrapf(err, "failed to connect to a TCP logging backend: %s", s.address)
	}
	s.log.Info("connected to TCP logging backend", "address", s.address)
	s.conn = conn
	return nil
}

func (s *sender) Send(record string) error {
	_, err := s.conn.Write(append([]byte(record), byte('\n')))
	return errors.Wrapf(err, "failed to send a log entry to a TCP logging backend: %s", s.address)
}

func (s *sender) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
