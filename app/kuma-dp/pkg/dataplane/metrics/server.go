package metrics

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"

	kumadp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

var logger = core.Log.WithName("metrics-hijacker")

var _ component.Component = &Hijacker{}

type Hijacker struct {
<<<<<<< HEAD
	envoyAdminPort uint32
	socketPath     string
}

func New(dataplane kumadp.Dataplane, envoyAdminPort uint32) *Hijacker {
	return &Hijacker{
		envoyAdminPort: envoyAdminPort,
		socketPath:     envoy.MetricsHijackerSocketName(dataplane.Name, dataplane.Mesh),
=======
	socketPath           string
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
		applicationsToScrape: applicationsToScrape,
>>>>>>> 7f1125714 (fix(*): do not override source address when TP is not enabled (#4951))
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
		"adminPort", s.envoyAdminPort,
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
func rewriteMetricsURL(port uint32, in *url.URL) string {
	u := url.URL{
		Scheme:   "http",
		Host:     fmt.Sprintf("127.0.0.1:%d", port),
		Path:     "/stats/prometheus",
		RawQuery: in.RawQuery,
	}

	return u.String()
}

func (s *Hijacker) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	resp, err := http.Get(rewriteMetricsURL(s.envoyAdminPort, req.URL))
	if err != nil {
<<<<<<< HEAD
		http.Error(writer, err.Error(), 500)
		return
=======
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
	if err != nil {
		logger.Error(err, "failed call", "name", app.Name, "path", app.Path, "port", app.Port)
		return nil
>>>>>>> 7f1125714 (fix(*): do not override source address when TP is not enabled (#4951))
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if err := MergeClusters(resp.Body, buf); err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}

	if _, err := writer.Write(buf.Bytes()); err != nil {
		logger.Error(err, "error while writing the response")
	}
}

func (s *Hijacker) NeedLeaderElection() bool {
	return false
}
