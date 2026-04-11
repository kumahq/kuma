package configfetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
)

type ConfigFetcher struct {
	httpClient        http.Client
	socketPath        string
	ticker            *time.Ticker
	handlers          []handlerInfo
	started           atomic.Bool
	perHandlerTimeout time.Duration
	firstStepDone     chan struct{}
	firstStepOnce     sync.Once
}

const unixDomainSocket = "unix"

var _ component.Component = &ConfigFetcher{}

var logger = core.Log.WithName("envoy-config-fetcher")

// NewConfigFetcher creates a new config Fetcher.
// Before calling Start we expect users to call ConfigFetcher.AddHandler for each configuration they want to refresh periodically
func NewConfigFetcher(socketPath string, ticker *time.Ticker, perHandlerTimeout time.Duration) *ConfigFetcher {
	return &ConfigFetcher{
		httpClient: http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, unixDomainSocket, socketPath)
				},
			},
		},
		handlers:          []handlerInfo{},
		socketPath:        socketPath,
		ticker:            ticker,
		perHandlerTimeout: perHandlerTimeout,
		firstStepDone:     make(chan struct{}),
	}
}

// FirstStepDone returns a channel that closes after the very first Step()
// has run all handlers, regardless of success. The kuma-dp readiness probe
// blocks on this so the application container isn't released by the
// wait-for-dataplane-ready PostStart hook before the embedded DNS proxy
// has had a chance to load mesh records from envoy admin /dns. Without
// this, the first request that hits a *.mesh hostname can fail with
// NXDOMAIN (see #16219).
func (cf *ConfigFetcher) FirstStepDone() <-chan struct{} {
	return cf.firstStepDone
}

type handlerInfo struct {
	path     string
	onChange OnHandlerChange
	lastEtag string
	metrics  *handlerMetrics
	l        logr.Logger
}

type OnHandlerChange func(ctx context.Context, reader io.Reader) error

// AddHandler add a Handler that will be polled for config change.
// this returns an error if the Handler.Path() doesn't have `/` prefix or if ConfigFetcher.Start was already called.
func (cf *ConfigFetcher) AddHandler(path string, onChange OnHandlerChange) error {
	if cf.started.Load() {
		return errors.New("can't add handler to config handler after startup")
	}
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("invalid path: %s, must start with '/'", path)
	}
	cf.handlers = append(cf.handlers, handlerInfo{
		path:     path,
		onChange: onChange,
		lastEtag: "",
		metrics:  newHandlerMetrics(path),
		l:        logger.WithValues("path", path),
	})
	return nil
}

// Start fetches sequentially from each handler calling Handler.OnChange if it changes.
// This expects the http server to return ETag headers and respect `If-None-Match` to avoid calling the Handler each time.
func (cf *ConfigFetcher) Start(stop <-chan struct{}) error {
	cf.started.Store(true)
	logger.Info("start",
		"socketPath", fmt.Sprintf("unix://%s", cf.socketPath),
	)
	defer logger.Info("stopped")

	// Step first to ensure we load conf ASAP
	if cf.Step() {
		cf.firstStepOnce.Do(func() { close(cf.firstStepDone) })
	}
	for {
		select {
		case <-cf.ticker.C:
			if cf.Step() {
				cf.firstStepOnce.Do(func() { close(cf.firstStepDone) })
			}
		case <-stop:
			return nil
		}
	}
}

func (cf *ConfigFetcher) NeedLeaderElection() bool {
	return false
}

// Step polls every registered handler once. It returns true if every
// handler reached envoy admin (got an HTTP response), even when that
// response is 404 or NotModified — only socket-not-exist / connection
// refused errors disqualify a handler. The return value drives the
// firstStepDone signal used by the readiness probe.
func (cf *ConfigFetcher) Step() bool {
	allReached := len(cf.handlers) > 0
	for i := range cf.handlers {
		h := &cf.handlers[i]
		h.metrics.HandlerTickCount.Add(1)
		start := time.Now()
		hasChanged, reached, err := cf.stepForHandler(h)
		if err != nil {
			h.metrics.HandlerErrorCount.Add(1)
			h.l.Error(err, "failed handle")
		}
		if !reached {
			allReached = false
		}
		if hasChanged { // Only compute duration when things changed
			h.metrics.HandlerTickDuration.Observe(time.Since(start).Seconds())
		}
	}
	return allReached
}

func (cf *ConfigFetcher) stepForHandler(h *handlerInfo) (bool, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cf.perHandlerTimeout)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://localhost%s", h.path), http.NoBody)
	if err != nil {
		return false, false, fmt.Errorf("failed to build request: %w", err)
	}
	if h.lastEtag != "" {
		r.Header.Set("If-None-Match", h.lastEtag)
	}
	response, err := cf.httpClient.Do(r) // #nosec G704 -- localhost socket communication, not user-controlled URL
	if err != nil {
		var opErr *net.OpError
		if errors.As(err, &opErr) && os.IsNotExist(opErr.Err) {
			h.l.V(1).Info("skipping fetch endpoint scrape since socket does not exist, this is likely about to start", "err", err)
			return false, false, nil
		}
		// this error can only occur when we configured policy once and then remove it. Listener is removed but socket file
		// is still present since Envoy does not clean it.
		if strings.Contains(err.Error(), "connection refused") {
			h.l.Info("failed to scrape config, Envoy not listening on socket")
			return false, false, nil
		}
		return false, false, fmt.Errorf("failed to scrape config: %w", err)
	}
	defer response.Body.Close()
	prevEtag := h.lastEtag
	h.lastEtag = "" // reset ETag to force refetch when errors happen
	switch response.StatusCode {
	case http.StatusOK:
		eTag := response.Header.Get("ETag")
		h.l.Info("scraped config from Envoy changed", "etag", eTag, "lastEtag", prevEtag)
		err = h.onChange(ctx, response.Body)
		if err == nil {
			h.lastEtag = eTag // only update ETag if we successfully processed the config
		}
		return true, true, err
	case http.StatusNotFound:
		h.l.V(1).Info("config scraped from Envoy is not found")
		return false, true, nil
	case http.StatusNotModified:
		h.l.V(1).Info("no change in config from Envoy")
		h.lastEtag = prevEtag
		return false, true, nil
	default:
		return false, false, fmt.Errorf("failed to scrape config: unexpected status code: %d", response.StatusCode)
	}
}
