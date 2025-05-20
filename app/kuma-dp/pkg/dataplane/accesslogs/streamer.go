//go:build !windows

package accesslogs

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"sync"
	"syscall"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

var logger = core.Log.WithName("access-log-streamer")

var _ component.Component = &accessLogStreamer{}

// accessLogStreamer implements TCP access logging in Kuma.
// When TCP logging is configured, CP configures Envoy to log access logs to a pipe file.
// Format of such logs is "address:port;msg"
// Streamer then reads logs from the pipe and passes it to a TCP destination.
type accessLogStreamer struct {
	address string

	sync.RWMutex
	senders map[string]*logSender
}

func (s *accessLogStreamer) NeedLeaderElection() bool {
	return false
}

func NewAccessLogStreamer(socketName string) component.Component {
	return &accessLogStreamer{
		address: socketName,
		senders: map[string]*logSender{},
	}
}

func (s *accessLogStreamer) Start(stop <-chan struct{}) error {
	logger.Info("cleaning existing access log pipe", "file", s.address)
	err := os.Remove(s.address)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrapf(err, "error removing existing fifo %s", s.address)
	}
	logger.Info("creating access log pipe", "file", s.address)
	err = syscall.Mkfifo(s.address, 0o666)
	if err != nil {
		return errors.Wrapf(err, "error creating fifo %s", s.address)
	}
	fd, err := os.OpenFile(s.address, os.O_CREATE, os.ModeNamedPipe)
	if err != nil {
		return errors.Wrapf(err, "error opening fifo %s", s.address)
	}

	reader := bufio.NewReader(fd)

	defer func() {
		fd.Close()
		s.cleanup()
	}()

	logger.Info("starting log streamer", "file", s.address)
	errCh := make(chan error, 1)
	go func() {
		if err := s.streamAccessLogs(reader); err != nil {
			errCh <- err
		}
	}()
	select {
	case err := <-errCh:
		return errors.Wrap(err, "log streamer terminated with an error")
	case <-stop:
		logger.Info("stopping log streamer")
		return nil
	}
}

func (s *accessLogStreamer) cleanup() {
	s.Lock()
	defer s.Unlock()
	for _, sender := range s.senders {
		logger.Info("closing connection to the TCP log destination", "address", sender)
		if err := sender.close(); err != nil {
			logger.Error(err, "could not close access log destination")
		}
	}
	s.senders = map[string]*logSender{}
}

func (s *accessLogStreamer) streamAccessLogs(reader *bufio.Reader) error {
	for {
		msg, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		var address string
		var accessLogMsg []byte

		type wrappedMessage struct {
			Address string      `json:"address"`
			Message interface{} `json:"message"`
		}

		var wrappedMsg wrappedMessage
		if err := json.Unmarshal(msg, &wrappedMsg); err != nil {
			// This is for compatibility if using TrafficLog
			parts := bytes.SplitN(msg, []byte(";"), 2)
			if len(parts) != 2 {
				logger.Error(nil, "log format invalid: expected 2 components separated by ';' or a JSON object")
				continue
			}

			address, accessLogMsg = string(parts[0]), parts[1]
		} else {
			address = wrappedMsg.Address

			if embeddedString, ok := wrappedMsg.Message.(string); ok {
				accessLogMsg = []byte(embeddedString)
			} else {
				accessLogMsg, err = json.Marshal(wrappedMsg.Message)
				if err != nil {
					logger.Error(err, "unable to marshal embedded message")
					continue
				}
				accessLogMsg = append(accessLogMsg, '\n')
			}
		}

		s.RLock()
		sender, initialized := s.senders[address]
		s.RUnlock()
		log := logger.WithValues("address", address)

		if !initialized {
			sender = &logSender{address: address}
			if err := sender.connect(); err != nil {
				// Drop log rather than return an error. Returning an error will cause reopening pipe which is unnecessary.
				// Do not retry this operation. If we were to retry here, the fifo can quickly grow.
				// Additionally, if TCP address is misconfigured, we would produce a lot of logs with misconfigured IP:port.
				// If we add retry, we would retry connection for every misconfigured log as the TCP destination is added to every log.
				// In this case, recovering logging to a proper configuration would take a lot of time.
				log.Error(err, "could not connect to TCP log destination. Dropping the log")
				continue
			}
			s.Lock()
			s.senders[address] = sender
			s.Unlock()
			log.Info("connected to TCP log destination")
		}

		if err := sender.send(accessLogMsg); err != nil {
			// If there is a problem on this connection, we need to reconnect.
			// Drop log rather than return an error. Returning an error would cause reopening pipe which is unnecessary.
			log.Error(err, "could not send the log to TCP log destination. Dropping the log", "address", address)
			if err := sender.close(); err != nil {
				log.Error(err, "could not close access log destination")
			}
			s.Lock()
			delete(s.senders, address)
			s.Unlock()
		}
	}
}
