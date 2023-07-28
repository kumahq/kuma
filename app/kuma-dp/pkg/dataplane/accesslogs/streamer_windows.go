// go:build windows
package accesslogs

import (
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

var logger = core.Log.WithName("access-log-streamer")

var _ component.Component = &accessLogStreamer{}

type accessLogStreamer struct {
	address string
}

func (s *accessLogStreamer) NeedLeaderElection() bool {
	return false
}

func NewAccessLogStreamer(address string) *accessLogStreamer {
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
