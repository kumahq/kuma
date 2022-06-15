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

type QueryParametersAppender func(in *url.URL) string

func EmptyQueryParametersAppender(url *url.URL) string {
	return ""
}

func EnvoyQueryParametersAppender(url *url.URL) string {
	queryParams := url.Query()
	queryParams.Add("format", "prometheus")
	return queryParams.Encode()
}

type ApplicationToScrape struct {
	Name          string
	Path          string
	Port          uint32
	QueryAppender QueryParametersAppender
	Mutator       MetricsMutator
}

type Hijacker struct {
	socketPath           string
	httpClient           http.Client
	applicationsToScrape []ApplicationToScrape
}

func New(dataplane kumadp.Dataplane, applicationsToScrape []ApplicationToScrape) *Hijacker {
	return &Hijacker{
		socketPath:           envoy.MetricsHijackerSocketName(dataplane.Name, dataplane.Mesh),
		httpClient:           http.Client{Timeout: 10 * time.Second},
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

// Depends on the specific application we do QueryParameters passing.
// Currently, we only support QueryParameters for Envoy metrics.
func rewriteMetricsURL(path string, port uint32, queryAppender QueryParametersAppender, in *url.URL) string {
	u := url.URL{
		Scheme:   "http",
		Host:     fmt.Sprintf("127.0.0.1:%d", port),
		Path:     path,
		RawQuery: queryAppender(in),
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

func (s *Hijacker) getStats(url *url.URL, app ApplicationToScrape) []byte {
	resp, err := s.httpClient.Get(rewriteMetricsURL(app.Path, app.Port, app.QueryAppender, url))
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
