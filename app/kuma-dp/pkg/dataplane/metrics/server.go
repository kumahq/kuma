package metrics

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"

	kumadp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

var logger = core.Log.WithName("metrics-hijacker")

var _ component.Component = &Hijacker{}

type MetricsMutator func(in io.Reader, out io.Writer) error

type ApplicationToScrape struct {
	Name    string
	Path    string
	Port    uint32
	Mutator MetricsMutator
}

type Hijacker struct {
	socketPath           string
<<<<<<< HEAD
	httpClient           http.Client
	applicationsToScrape []ApplicationToScrape
}

func New(dataplane kumadp.Dataplane, applicationsToScrape []ApplicationToScrape) *Hijacker {
	return &Hijacker{
		socketPath:           envoy.MetricsHijackerSocketName(dataplane.Name, dataplane.Mesh),
		httpClient:           http.Client{Timeout: 10 * time.Second},
=======
	httpClientIPv4       http.Client
	httpClientIPv6       http.Client
	applicationsToScrape []ApplicationToScrape
}

func createHttpClient(isUsingTransparentProxy bool, sourceAddress *net.TCPAddr) http.Client {
	// we need this in case of not localhost requests, it returns fast in iptabels
	if isUsingTransparentProxy {
		dialer := &net.Dialer{
			LocalAddr: sourceAddress,
		}
		return http.Client{
			Transport: &http.Transport{
				DialContext: dialer.DialContext,
			},
		}
	}
	return http.Client{}
}

func New(dataplane kumadp.Dataplane, applicationsToScrape []ApplicationToScrape, isUsingTransparentProxy bool) *Hijacker {
	return &Hijacker{
		socketPath:           envoy.MetricsHijackerSocketName(dataplane.Name, dataplane.Mesh),
		httpClientIPv4:       createHttpClient(isUsingTransparentProxy, inPassThroughIPv4),
		httpClientIPv6:       createHttpClient(isUsingTransparentProxy, inPassThroughIPv6),
>>>>>>> 7f1125714 (fix(*): do not override source address when TP is not enabled (#4951))
		applicationsToScrape: applicationsToScrape,
	}
}

func (s *Hijacker) Start(stop <-chan struct{}) error {
	_, err := os.Stat(s.socketPath)
	if err == nil {
		// File is accessible try to rename it to verify it is not open
		newName := s.socketPath + ".bak"
		err := os.Rename(s.socketPath, newName)
		if err != nil {
			return errors.Errorf("file %s exists and probably opened by another kuma-dp instance", s.socketPath)
		}
		err = os.Remove(newName)
		if err != nil {
			return errors.Errorf("not able the delete the backup file %s", newName)
		}
	}

	lis, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return err
	}

	defer func() {
		lis.Close()
	}()

	logger.Info("starting Metrics Hijacker Server",
		"socketPath", fmt.Sprintf("unix://%s", s.socketPath),
	)

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

// The Envoy stats endpoint recognizes the "used_only" and "filter" query
// parameters. We squash the path to enforce Prometheus metrics format, but
// forward the query parameters so that the scraper can do partial scrapes.
func rewriteMetricsURL(path string, port uint32, in *url.URL) string {
	u := url.URL{
		Scheme:   "http",
		Host:     fmt.Sprintf("127.0.0.1:%d", port),
		Path:     path,
		RawQuery: in.RawQuery,
	}

	return u.String()
}

func (s *Hijacker) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	out := make(chan []byte, len(s.applicationsToScrape))
	var wg sync.WaitGroup
	wg.Add(len(s.applicationsToScrape))
	go func() {
		wg.Wait()
		close(out)
	}()

	for _, app := range s.applicationsToScrape {
		go func(app ApplicationToScrape) {
			defer wg.Done()
			out <- s.getStats(req.URL, app)
		}(app)
	}
	for resp := range out {
		if _, err := writer.Write(resp); err != nil {
			logger.Error(err, "error while writing the response")
		}
		if _, err := writer.Write([]byte("\n")); err != nil {
			logger.Error(err, "error while writing the response")
		}
	}
}

<<<<<<< HEAD
func (s *Hijacker) getStats(url *url.URL, app ApplicationToScrape) []byte {
	resp, err := s.httpClient.Get(rewriteMetricsURL(app.Path, app.Port, url))
=======
func (s *Hijacker) getStats(ctx context.Context, initReq *http.Request, app ApplicationToScrape) []byte {
	req, err := http.NewRequest("GET", rewriteMetricsURL(app.Address, app.Port, app.Path, app.QueryModifier, initReq.URL), nil)
	if err != nil {
		logger.Error(err, "failed to create request")
		return nil
	}
	s.passRequestHeaders(req.Header, initReq.Header)
	req = req.WithContext(ctx)
	var resp *http.Response
	if app.IsIPv6 {
		resp, err = s.httpClientIPv6.Do(req)
		if err == nil {
			defer resp.Body.Close()
		}
	} else {
		resp, err = s.httpClientIPv4.Do(req)
		if err == nil {
			defer resp.Body.Close()
		}
	}
>>>>>>> 7f1125714 (fix(*): do not override source address when TP is not enabled (#4951))
	if err != nil {
		logger.Error(err, "failed call", "name", app.Name, "path", app.Path, "port", app.Port)
		return nil
	}
	defer resp.Body.Close()

	var bodyBytes []byte
	if app.Mutator != nil {
		buf := new(bytes.Buffer)
		if err := app.Mutator(resp.Body, buf); err != nil {
			logger.Error(err, "failed while mutating data", "name", app.Name, "path", app.Path, "port", app.Port)
			return nil
		}
		bodyBytes = buf.Bytes()
	} else {
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			logger.Error(err, "failed while writing", "name", app.Name, "path", app.Path, "port", app.Port)
			return nil
		}
	}
	return bodyBytes
}

func (s *Hijacker) NeedLeaderElection() bool {
	return false
}
