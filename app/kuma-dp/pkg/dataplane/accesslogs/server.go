package accesslogs

import (
	"bufio"
	"io"

	"github.com/pkg/errors"

	v3 "github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/accesslogs/v3"
	kumadp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

var logger = core.Log.WithName("accesslogs-server")

var _ component.Component = &accessLogServer{}

type accessLogServer struct {
	address string
}

func (s *accessLogServer) NeedLeaderElection() bool {
	return false
}

func NewAccessLogServer(dataplane kumadp.Dataplane) *accessLogServer {
	address := envoy.AccessLogSocketName(dataplane.Name, dataplane.Mesh)
	return &accessLogServer{
		address: address,
	}
}

type ReaderCloser interface {
	io.Reader
	io.Closer
}

func (s *accessLogServer) Start(stop <-chan struct{}) error {
	fd, err := streamer(s.address)
	if err != nil {
		return errors.Wrap(err, "error creating reader")
	}

	reader := bufio.NewReader(fd)

	defer func() {
		fd.Close()
	}()

	logger.Info("starting Access Log Server", "pipefile", s.address)
	errCh := make(chan error, 1)

	alStreamer := v3.NewAccessLogStreamer()
	go func() {
		if err := alStreamer.StreamAccessLogs(reader); err != nil {
			errCh <- err
		}
	}()
	select {
	case err := <-errCh:
		return err
	case <-stop:
		logger.Info("stopping Access Log Server")
		return nil
	}
}
