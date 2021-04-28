package metrics

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

var logger = core.Log.WithName("metrics-hijacker")

var _ component.Component = &Hijacker{}

type Hijacker struct {
	envoyAdminPort uint32
	address        string
}

func New(envoyAdminPort uint32) *Hijacker {
	return &Hijacker{
		envoyAdminPort: envoyAdminPort,
		address:        envoy.MetricsHijackerSocketName(),
	}
}

func (s *Hijacker) Start(stop <-chan struct{}) error {
	_, err := os.Stat(s.address)
	if err == nil {
		// File is accessible try to rename it to verify it is not open
		newName := s.address + ".bak"
		err := os.Rename(s.address, newName)
		if err != nil {
			return errors.Errorf("file %s exists and probably opened by another kuma-dp instance", s.address)
		}
		err = os.Remove(newName)
		if err != nil {
			return errors.Errorf("not able the delete the backup file %s", newName)
		}
	}

	lis, err := net.Listen("unix", s.address)
	if err != nil {
		return err
	}

	defer func() {
		lis.Close()
	}()

	logger.Info("starting Metrics Hijacker Server", "address", fmt.Sprintf("unix://%s", s.address))

	server := &http.Server{
		Handler: s,
	}

	errCh := make(chan error)
	go func() {
		if err := server.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-stop:
		logger.Info("stopping Metrics Hijacker Server")
		return server.Shutdown(context.Background())
	}
}

func (s *Hijacker) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/stats/prometheus", s.envoyAdminPort))
	if err != nil {
		if _, err := writer.Write([]byte(err.Error())); err != nil {
			logger.Error(err, "error while writing the response")
		}
		return
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		if _, err := writer.Write([]byte(err.Error())); err != nil {
			logger.Error(err, "error while writing the response")
		}
		return
	}
	if _, err := writer.Write(buf.Bytes()); err != nil {
		logger.Error(err, "error while writing the response")
	}
}

func (s *Hijacker) NeedLeaderElection() bool {
	return false
}
