// go:build windows
package accesslogs

import (
	kumadp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

var logger = core.Log.WithName("access-log-streamer")

var _ component.Component = &accessLogStreamer{}

type accessLogStreamer struct {
	address string
}

func (s *accessLogStreamer) NeedLeaderElection() bool {
	return false
}

func NewAccessLogStreamer(dataplane kumadp.Dataplane) *accessLogStreamer {
	address := envoy.AccessLogSocketName(dataplane.Name, dataplane.Mesh)
	return &accessLogStreamer{
		address: address,
	}
}

func (s *accessLogStreamer) Start(stop <-chan struct{}) error {
	logger.Info("TCP AccessLog is not supported on windows")
	select {
	case <-stop:
		logger.Info("stopping Access Log Server")
		return nil
	}
}
