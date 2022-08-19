package accesslogs

import (
	"net"
	"time"

	"github.com/pkg/errors"
)

const (
	defaultConnectTimeout = 5 * time.Second
)

type logSender struct {
	address string
	conn    net.Conn
}

func (s *logSender) connect() error {
	conn, err := net.DialTimeout("tcp", s.address, defaultConnectTimeout)
	if err != nil {
		return errors.Wrapf(err, "failed to connect to a TCP logging backend: %s", s.address)
	}
	s.conn = conn
	return nil
}

func (s *logSender) send(record []byte) error {
	if s.conn == nil {
		return errors.New("connection not initialized")
	}
	_, err := s.conn.Write(record)
	return errors.Wrapf(err, "failed to send a log entry to a TCP logging backend: %s", s.address)
}

func (s *logSender) close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
