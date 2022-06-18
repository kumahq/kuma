package accesslogs

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"

	v3 "github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/accesslogs/v3"
	kumadp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

var logger = core.Log.WithName("accesslogs-server")

var _ component.Component = &accessLogServer{}

type accessLogServer struct {
	server  *grpc.Server
	address string
}

func (s *accessLogServer) NeedLeaderElection() bool {
	return false
}

func NewAccessLogServer(dataplane kumadp.Dataplane) *accessLogServer {
	address := envoy.AccessLogSocketName(dataplane.Name, dataplane.Mesh)
	return &accessLogServer{
		server:  grpc.NewServer(),
		address: address,
	}
}

func (s *accessLogServer) Start(stop <-chan struct{}) error {
	v3.RegisterAccessLogServer(s.server)

	_, err := os.Stat(s.address)
	if err == nil {
		// File is accessible try to rename it to verify it is not open
		newName := s.address + ".bak"
		err := os.Rename(s.address, newName)
		if err != nil {
			return fmt.Errorf("file %s exists and probably opened by another kuma-dp instance", s.address)
		}
		err = os.Remove(newName)
		if err != nil {
			return fmt.Errorf("not able the delete the backup file %s", newName)
		}
	}

	lis, err := net.Listen("unix", s.address)
	if err != nil {
		return err
	}

	defer func() {
		lis.Close()
	}()

	logger.Info("starting Access Log Server", "address", fmt.Sprintf("unix://%s", s.address))
	errCh := make(chan error, 1)
	go func() {
		if err := s.server.Serve(lis); err != nil {
			errCh <- err
		}
	}()
	select {
	case err := <-errCh:
		return err
	case <-stop:
		logger.Info("stopping Access Log Server")
		s.server.GracefulStop()
		return nil
	}
}
