package readiness

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/bakito/go-log-logr-adapter/adapter"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

const (
	pathPrefixReady   = "/ready"
	stateReady        = "READY"
	stateInitializing = "INITIALIZING"
	stateDraining     = "DRAINING"
)

// Reporter reports the health status of this Kuma Dataplane Proxy
type Reporter struct {
	socketPath string
	envoyProbe *EnvoyReadinessProbe
	draining   atomic.Bool
}

var logger = core.Log.WithName("readiness")

func NewReporter(socketPath string, adminAddress string, adminPort uint32) *Reporter {
	var envoyProbe *EnvoyReadinessProbe
	if adminPort > 0 {
		envoyProbe = &EnvoyReadinessProbe{
			LocalHostAddr: adminAddress,
			AdminPort:     uint16(adminPort),
		}
	}

	return &Reporter{
		socketPath: socketPath,
		envoyProbe: envoyProbe,
	}
}

func (r *Reporter) Start(stop <-chan struct{}) error {
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
	mux.HandleFunc(pathPrefixReady, r.handleReadiness)
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

func (r *Reporter) Draining() {
	r.draining.Store(true)
}

func (r *Reporter) handleReadiness(writer http.ResponseWriter, req *http.Request) {
	state := stateInitializing
	var envoyReadinessErr error
	if r.envoyProbe != nil {
		envoyReadinessErr = r.envoyProbe.Check()
	}

	if envoyReadinessErr == nil {
		draining := r.draining.Load()
		if !draining {
			state = stateReady
		} else {
			state = stateDraining
		}
	} else {
		logger.Info("[WARNING] Envoy is not ready", "err", envoyReadinessErr.Error())
	}

	stateBytes := []byte(state + "\n")
	writer.Header().Set("Content-Type", "text/plain")
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(stateBytes)))
	writer.Header().Set("Cache-Control", "no-cache, max-age=0")
	writer.Header().Set("X-Readiness-Server", "kuma-dp")
	stateHTTPStatus := http.StatusServiceUnavailable
	if state == stateReady {
		stateHTTPStatus = http.StatusOK
	}
	writer.WriteHeader(stateHTTPStatus)
	_, err := writer.Write(stateBytes)
	logger.V(1).Info("responding readiness state", "state", state, "client", req.RemoteAddr)
	if err != nil {
		logger.Info("[WARNING] could not write response", "err", err)
	}
}

func (r *Reporter) NeedLeaderElection() bool {
	return false
}

var _ component.Component = &Reporter{}
