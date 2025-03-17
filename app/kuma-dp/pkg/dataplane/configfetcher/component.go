package configfetcher

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

type ConfigFetcher struct {
	httpClient        http.Client
	socketPath        string
	ticker            *time.Ticker
	handlers          []handlerInfo
	started           atomic.Bool
	perHandlerTimeout time.Duration
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
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial(unixDomainSocket, socketPath)
				},
			},
		},
		handlers:          []handlerInfo{},
		socketPath:        socketPath,
		ticker:            ticker,
		perHandlerTimeout: perHandlerTimeout,
	}
}

type handlerInfo struct {
	lastEtag string
	handler  Handler
	metrics  *handlerMetrics
	l        logr.Logger
}

// AddHandler add a Handler that will be polled for config change.
// this returns an error if the Handler.Path() doesn't have `/` prefix or if ConfigFetcher.Start was already called.
func (cf *ConfigFetcher) AddHandler(h Handler) error {
	if cf.started.Load() {
		return errors.New("can't add handler to config handler after startup")
	}
	if !strings.HasPrefix(h.Path(), "/") {
		return fmt.Errorf("invalid path: %s, must start with '/'", h.Path())
	}
	cf.handlers = append(cf.handlers, handlerInfo{
		handler:  h,
		lastEtag: "",
		metrics:  newHandlerMetrics(h.Path()),
		l:        logger.WithValues("path", h.Path()),
	})
	return nil
}

// Start fetches sequentially from each handler calling Handler.OnChange if it changes.
// This expects the http server to return ETag headers and respect `If-None-Match` to avoid calling the Handler each time.
func (cf *ConfigFetcher) Start(stop <-chan struct{}) error {
	cf.started.Store(true)
	logger.Info("starting configuration scraper",
		"socketPath", fmt.Sprintf("unix://%s", cf.socketPath),
	)
	defer logger.Info("configuration scraper stopped")

	// Step first to ensure we load conf ASAP
	cf.Step()
	for {
		select {
		case <-cf.ticker.C:
			cf.Step()
		case <-stop:
			logger.Info("starting shutdown")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			for _, h := range cf.handlers {
				h.metrics.HandlerShutdownCount.Add(1)
				err := h.handler.Shutdown(ctx)
				if err != nil {
					logger.WithValues("path", h.handler.Path()).Error(err, "failed to shutdown")
					h.metrics.HandlerErrorCount.Add(1)
				}
			}
			return nil
		}
	}
}

func (cf *ConfigFetcher) NeedLeaderElection() bool {
	return false
}

func (cf *ConfigFetcher) Step() {
	if _, err := os.Stat(cf.socketPath); errors.Is(err, os.ErrNotExist) {
		logger.Info("skipping fetch endpoint scrape since socket does not exist, this is likely about to start", "err", err)
		return
	}
	for i := range cf.handlers {
		h := &cf.handlers[i]
		h.metrics.HandlerTickCount.Add(1)
		start := time.Now()
		hasChanged, err := cf.stepForHandler(h)
		if err != nil {
			h.metrics.HandlerErrorCount.Add(1)
			h.l.Error(err, "failed handle")
		}
		if hasChanged { // Only compute duration when things changed
			h.metrics.HandlerTickDuration.Observe(time.Since(start).Seconds())
		}
	}
}

func (cf *ConfigFetcher) stepForHandler(h *handlerInfo) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cf.perHandlerTimeout)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://localhost%s", h.handler.Path()), nil)
	if err != nil {
		return false, fmt.Errorf("failed to build request: %w", err)
	}
	if h.lastEtag != "" {
		r.Header.Set("If-None-Match", h.lastEtag)
	}
	response, err := cf.httpClient.Do(r)
	if err != nil {
		// this error can only occur when we configured policy once and then remove it. Listener is removed but socket file
		// is still present since Envoy does not clean it.
		if strings.Contains(err.Error(), "connection refused") {
			h.l.Info("failed to scrape config, Envoy not listening on socket")
			return false, nil
		}
		return false, fmt.Errorf("failed to scrape config: %w", err)
	}
	defer response.Body.Close()
	prevEtag := h.lastEtag
	h.lastEtag = "" // reset ETag to force refetch when errors happen
	switch response.StatusCode {
	case http.StatusOK:
		etag := response.Header.Get("ETag")
		h.l.Info("scraped config from Envoy changed", "etag", etag, "lastEtag", prevEtag)
		h.lastEtag = etag
		err = h.handler.OnChange(ctx, response.Body)
		if err != nil {
			h.lastEtag = "" // reset ETag to force refetch when errors happen
		}
		return true, err
	case http.StatusNotFound:
		h.l.V(1).Info("config scraped from Envoy is not found")
		return false, nil
	case http.StatusNotModified:
		h.l.V(1).Info("no change in config from Envoy")
		return false, nil
	default:
		return false, fmt.Errorf("failed to scrape config: unexpected status code: %d", response.StatusCode)
	}
}
