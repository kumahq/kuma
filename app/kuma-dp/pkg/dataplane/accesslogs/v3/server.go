package v3

import (
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v3"
	"google.golang.org/grpc"

	"github.com/kumahq/kuma/pkg/core"
)

var logger = core.Log.WithName("accesslogs-server")

type accessLogServer struct {
	newHandler logHandlerFactoryFunc

	// streamCount for counting streams
	streamCount int64
}

var _ envoy_accesslog.AccessLogServiceServer = &accessLogServer{}

func RegisterAccessLogServer(server *grpc.Server) {
	srv := &accessLogServer{
		newHandler: defaultHandler,
	}
	envoy_accesslog.RegisterAccessLogServiceServer(server, srv)
}

func (s *accessLogServer) StreamAccessLogs(stream envoy_accesslog.AccessLogService_StreamAccessLogsServer) (err error) {
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
		msg, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		totalRequests++

		if !initialized {
			initialized = true

			handler, err = s.newHandler(log, msg)
			if err != nil {
				return fmt.Errorf("failed to initialize Access Logs stream: %w", err)
			}
			defer handler.Close()
		}

		if err := handler.Handle(msg); err != nil {
			return err
		}
	}
}
