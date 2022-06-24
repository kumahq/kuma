package v3

import (
	"bufio"
	"bytes"
	"io"
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
)

var logger = core.Log.WithName("accesslogs-server")

type accessLogStreamer struct {
	newHandler logHandlerFactoryFunc

	// streamCount for counting streams
	streamCount int64
}

func NewAccessLogStreamer() *accessLogStreamer {
	srv := &accessLogStreamer{
		newHandler: defaultHandler,
	}
	return srv
}

func (s *accessLogStreamer) StreamAccessLogs(reader *bufio.Reader) (err error) {
	// increment stream count
	streamID := atomic.AddInt64(&s.streamCount, 1)

	totalRequests := 0

	log := logger.WithValues("streamID", streamID)
	log.Info("starting a new Access Logs stream")
	defer func() {
		log.Info("Access Logs stream is terminated", "totalRequests", totalRequests)
		if err != nil {
			log.Error(err, "Access Logs stream terminated with an error")
		}
	}()

	initialized := false
	var handler logHandler
	for {
		msg, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		totalRequests++

		parts := bytes.SplitN(msg, []byte(";"), 2)
		if len(parts) != 2 {
			log.Error(nil, "log format invalid expected 2 components separated by ';'", "ncomponents", len(parts))
			continue
		}
		address, accessLogMsg := string(parts[0]), parts[1]

		if !initialized {
			initialized = true

			handler, err = s.newHandler(log, address)
			if err != nil {
				return errors.Wrap(err, "failed to initialize Access Logs stream")
			}
			defer handler.Close()
		}

		if err := handler.Handle(accessLogMsg); err != nil {
			return err
		}
	}
}
