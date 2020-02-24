package accesslogs

import (
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"google.golang.org/grpc"

	kumadp "github.com/Kong/kuma/pkg/config/app/kuma-dp"
	"github.com/Kong/kuma/pkg/core"
	"github.com/pkg/errors"

	envoy_accesslog "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v2"
)

var logger = core.Log.WithName("accesslogs-server")

type accessLogServer struct {
	server     *grpc.Server
	newHandler logHandlerFactoryFunc

	// streamCount for counting streams
	streamCount int64
}

func NewAccessLogServer() *accessLogServer {
	return &accessLogServer{
		server:     grpc.NewServer(),
		newHandler: defaultHandler,
	}
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
			if err == io.EOF {
				return nil
			}
			return err
		}

		totalRequests++

		if !initialized {
			initialized = true

			handler, err = s.newHandler(log, msg)
			if err != nil {
				return errors.Wrap(err, "failed to initialize Access Logs stream")
			}
			defer handler.Close()
		}

		if err := handler.Handle(msg); err != nil {
			return err
		}
	}
}

func (s *accessLogServer) Start(dataplane kumadp.Dataplane) error {
	envoy_accesslog.RegisterAccessLogServiceServer(s.server, s)
	address := fmt.Sprintf("/tmp/kuma-access-logs-%s-%s.sock", dataplane.Name, dataplane.Mesh)
	lis, err := net.Listen("unix", address)
	if err != nil {
		return err
	}
	logger.Info("starting", "address", fmt.Sprintf("unix://%s", address))
	if err := s.server.Serve(lis); err != nil {
		logger.Error(err, "terminated with an error")
		return err
	}
	return nil
}

func (s *accessLogServer) Close() {
	s.server.GracefulStop()
}

var _ envoy_accesslog.AccessLogServiceServer = &accessLogServer{}
