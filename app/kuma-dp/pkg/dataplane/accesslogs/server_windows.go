// go:build windows
package accesslogs

import (
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

func (s *accessLogServer) Start(stop <-chan struct{}) error {
	logger.Info("TCP AccessLog is not supported on windows")
	select {
	case <-stop:
		logger.Info("stopping Access Log Server")
		return nil
	}
}
