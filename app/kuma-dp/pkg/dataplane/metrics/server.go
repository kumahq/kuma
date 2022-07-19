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

	"github.com/pkg/errors"

	kumadp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

var logger = core.Log.WithName("metrics-hijacker")

var _ component.Component = &Hijacker{}

type MetricsMutator func(in io.Reader, out io.Writer) error

type QueryParametersModifier func(queryParameters url.Values) string

func RemoveQueryParameters(_ url.Values) string {
	return ""
}

func AddPrometheusFormat(queryParameters url.Values) string {
	queryParameters.Add("format", "prometheus")
	return queryParameters.Encode()
}

type ApplicationToScrape struct {
	Name          string
	Path          string
	Port          uint32
	QueryModifier QueryParametersModifier
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
		httpClient:           http.Client{},
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

// We pass QueryParameters only for the specific application.
// Currently, we only support QueryParameters for Envoy metrics.
func rewriteMetricsURL(path string, port uint32, queryModifier QueryParametersModifier, in *url.URL) string {
	u := url.URL{
		Scheme:   "http",
		Host:     fmt.Sprintf("127.0.0.1:%d", port),
		Path:     path,
		RawQuery: queryModifier(in.Query()),
	}
	return u.String()
}

func (s *Hijacker) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	out := make(chan []byte, len(s.applicationsToScrape))
	var wg sync.WaitGroup
	done := make(chan []byte)
	wg.Add(len(s.applicationsToScrape))
	go func() {
		wg.Wait()
		close(done)
		close(out)
	}()

	for _, app := range s.applicationsToScrape {
		go func(app ApplicationToScrape) {
			defer wg.Done()
			out <- s.getStats(req.URL, app, ctx)
		}(app)
	}

	select {
	case <-ctx.Done():
		return
	case <-done:
		for resp := range out {
			if _, err := writer.Write(resp); err != nil {
				logger.Error(err, "error while writing the response")
			}
			if _, err := writer.Write([]byte("\n")); err != nil {
				logger.Error(err, "error while writing the response")
			}
		}
	}
}

func (s *Hijacker) getStats(url *url.URL, app ApplicationToScrape, context context.Context) []byte {
	req, _ := http.NewRequest("GET", rewriteMetricsURL(app.Path, app.Port, app.QueryModifier, url), nil)
	req = req.WithContext(context)
	resp, err := s.httpClient.Do(req)
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
