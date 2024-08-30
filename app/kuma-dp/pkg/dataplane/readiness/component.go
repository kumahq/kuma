package readiness

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/bakito/go-log-logr-adapter/adapter"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

const (
	pathPrefixReady  = "/ready"
	stateReady       = "READY"
	stateTerminating = "TERMINATING"
)

// Reporter reports the health status of this Kuma Dataplane Proxy
type Reporter struct {
	localListenAddr string
	localListenPort uint32
	isTerminating   atomic.Bool
}

var logger = core.Log.WithName("readiness")

func NewReporter(localIPAddr string, localListenPort uint32) *Reporter {
	return &Reporter{
		localListenPort: localListenPort,
		localListenAddr: localIPAddr,
	}
}

func (r *Reporter) Start(stop <-chan struct{}) error {
	protocol := "tcp"
	addr := r.localListenAddr
	if govalidator.IsIPv6(addr) {
		protocol = "tcp6"
		addr = fmt.Sprintf("[%s]", addr)
	}
	lis, err := net.Listen(protocol, fmt.Sprintf("%s:%d", addr, r.localListenPort))
	if err != nil {
		return err
	}

	defer func() {
		_ = lis.Close()
	}()

	logger.Info("starting readiness reporter", "addr", lis.Addr().String())

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

func (r *Reporter) Terminating() {
	r.isTerminating.Store(true)
}

func (r *Reporter) handleReadiness(writer http.ResponseWriter, req *http.Request) {
	state := stateReady
	stateHTTPStatus := http.StatusOK
	if r.isTerminating.Load() {
		state = stateTerminating
		stateHTTPStatus = http.StatusServiceUnavailable
	}

	stateBytes := []byte(state)
	writer.Header().Set("content-type", "text/plain")
	writer.Header().Set("content-length", fmt.Sprintf("%d", len(stateBytes)))
	writer.Header().Set("cache-control", "no-cache, max-age=0")
	writer.Header().Set("x-powered-by", "kuma-dp")
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
