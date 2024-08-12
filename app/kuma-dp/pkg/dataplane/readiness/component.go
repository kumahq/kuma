package readiness

import (
	"context"
	"fmt"
	"github.com/bakito/go-log-logr-adapter/adapter"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

const (
	PathPrefixReady = "/ready"
)

// reporter reports the health status of this Kuma Dataplane Proxy
// it always responds with "READY" currently
type reporter struct {
	socketPath string
}

var logger = core.Log.WithName("readiness")

func NewReporter(socketPath string) component.Component {
	return &reporter{
		socketPath: socketPath,
	}
}

func (r *reporter) Start(stop <-chan struct{}) error {
	_, err := os.Stat(r.socketPath)
	if err == nil {
		// File is accessible try to rename it to verify it is not open
		newName := r.socketPath + ".bak"
		err = os.Rename(r.socketPath, newName)
		if err != nil {
			return errors.Errorf("file %s exists and probably opened by another kuma-dp instance", r.socketPath)
		}
		err = os.Remove(newName)
		if err != nil {
			return errors.Errorf("not able the delete the backup file %s", newName)
		}
	}

	lis, err := net.Listen("unix", r.socketPath)
	if err != nil {
		return err
	}

	defer func() {
		_ = lis.Close()
	}()

	logger.Info("starting readiness reporter",
		"socketPath", fmt.Sprintf("unix://%s", r.socketPath),
	)

	mux := http.NewServeMux()
	mux.HandleFunc(PathPrefixReady, r.handleReadiness)
	server := &http.Server{
		ReadHeaderTimeout: time.Second,
		Handler:           mux,
		ErrorLog:          adapter.ToStd(logger),
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
		logger.Info("stopping readiness reporter")
		return server.Shutdown(context.Background())
	}
}

func (r *reporter) handleReadiness(writer http.ResponseWriter, _ *http.Request) {
	readyBytes := []byte("READY")
	writer.Header().Set("Content-Type", "text/plain")
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(readyBytes)))
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write(readyBytes)
	if err != nil {
		logger.Info("[WARNING] could not write response", "err", err)
	}
}

func (r *reporter) NeedLeaderElection() bool {
	return false
}

var _ component.Component = &reporter{}
