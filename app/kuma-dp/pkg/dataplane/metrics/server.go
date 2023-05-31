package metrics

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/common/expfmt"

	kumadp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

var (
	inPassThroughIPv4 = &net.TCPAddr{IP: net.ParseIP("127.0.0.6")}
	inPassThroughIPv6 = &net.TCPAddr{IP: net.ParseIP("::6")}
)

var (
	prometheusRequestHeaders = []string{"accept", "accept-encoding", "user-agent", "x-prometheus-scrape-timeout-seconds"}
	logger                   = core.Log.WithName("metrics-hijacker")

	// holds prometheus content types in order of priority.
	prometheusPriorityContentType = []expfmt.Format{
		expfmt.FmtOpenMetrics_1_0_0,
		expfmt.FmtOpenMetrics_0_0_1,
		expfmt.FmtText,
		expfmt.FmtUnknown,
	}

	// Reverse mapping of prometheusPriorityContentType for faster lookup.
	prometheusPriorityContentTypeLookup = func(expformats []expfmt.Format) map[expfmt.Format]int32 {
		reverseMapping := map[expfmt.Format]int32{}
		for priority, format := range expformats {
			reverseMapping[format] = int32(priority)
		}
		return reverseMapping
	}(prometheusPriorityContentType)
)

var _ component.Component = &Hijacker{}

type MetricsMutator func(in io.Reader, out io.Writer) error

type QueryParametersModifier func(queryParameters url.Values) string

func RemoveQueryParameters(_ url.Values) string {
	return ""
}

func AddPrometheusFormat(queryParameters url.Values) string {
	queryParameters.Add("format", "prometheus")
	queryParameters.Add("text_readouts", "")
	return queryParameters.Encode()
}

type ApplicationToScrape struct {
	Name          string
	Address       string
	Path          string
	Port          uint32
	IsIPv6        bool
	QueryModifier QueryParametersModifier
	Mutator       MetricsMutator
}

type Hijacker struct {
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
		ReadHeaderTimeout: time.Second,
		Handler:           s,
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
func rewriteMetricsURL(address string, port uint32, path string, queryModifier QueryParametersModifier, in *url.URL) string {
	u := url.URL{
		Scheme:   "http",
		Host:     net.JoinHostPort(address, strconv.FormatUint(uint64(port), 10)),
		Path:     path,
		RawQuery: queryModifier(in.Query()),
	}
	return u.String()
}

func (s *Hijacker) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	out := make(chan []byte, len(s.applicationsToScrape))
	contentTypes := make(chan expfmt.Format, len(s.applicationsToScrape))
	var wg sync.WaitGroup
	done := make(chan []byte)
	wg.Add(len(s.applicationsToScrape))
	go func() {
		wg.Wait()
		close(out)
		close(contentTypes)
		close(done)
	}()

	for _, app := range s.applicationsToScrape {
		go func(app ApplicationToScrape) {
			defer wg.Done()
			content, contentType := s.getStats(ctx, req, app)
			out <- content

			// It's possible to track the highest priority content type seen,
			// but that would require mutex.
			// I would prefer to calculate it later at one go
			contentTypes <- contentType
		}(app)
	}

	select {
	case <-ctx.Done():
		return
	case <-done:
		writer.Header().Set(hdrContentType, string(selectContentType(contentTypes, req.Header)))
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

// selectContentType selects the highest priority content type supported by the applications.
// If no valid content type is returned by the applications, it returns the highest priority
// content type supported by the scraper.
func selectContentType(contentTypes <-chan expfmt.Format, reqHeader http.Header) expfmt.Format {
	// Tracks highest negotiated content type priority.
	// Lower number means higher priority
	//
	// We should not simply use the highest priority content type even if `application/openmetrics-text`
	// is the superset of `text/plain`, as it might not be
	// supported by the applications or the user might be using older prom scraper.
	// So it's better to choose the highest negotiated content type between the apps and the scraper.
	var ctPriority int32 = math.MaxInt32
	ct := expfmt.FmtUnknown
	for contentType := range contentTypes {
		priority, valid := prometheusPriorityContentTypeLookup[contentType]
		if !valid {
			continue
		}
		if priority < ctPriority {
			ctPriority = priority
			ct = contentType
		}
	}

	// If no valid content type is returned by the applications,
	// use the highest priority content type supported by the scraper.
	if ct == expfmt.FmtUnknown {
		ct = expfmt.NegotiateIncludingOpenMetrics(reqHeader)
	}

	return ct
}

func (s *Hijacker) getStats(ctx context.Context, initReq *http.Request, app ApplicationToScrape) ([]byte, expfmt.Format) {
	req, err := http.NewRequest("GET", rewriteMetricsURL(app.Address, app.Port, app.Path, app.QueryModifier, initReq.URL), nil)
	if err != nil {
		logger.Error(err, "failed to create request")
		return nil, ""
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
		return nil, ""
	}

	respContentType := responseFormat(resp.Header)

	var bodyBytes []byte
	if app.Mutator != nil {
		buf := new(bytes.Buffer)
		if err := app.Mutator(resp.Body, buf); err != nil {
			logger.Error(err, "failed while mutating data", "name", app.Name, "path", app.Path, "port", app.Port)
			return nil, ""
		}
		bodyBytes = buf.Bytes()
	} else {
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			logger.Error(err, "failed while writing", "name", app.Name, "path", app.Path, "port", app.Port)
			return nil, ""
		}
	}
	return bodyBytes, respContentType
}

func (s *Hijacker) passRequestHeaders(into http.Header, from http.Header) {
	// pass request headers
	// https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go#L772
	for _, header := range prometheusRequestHeaders {
		val := from.Get(header)
		if val != "" {
			into.Set(header, val)
		}
	}
}

func (s *Hijacker) NeedLeaderElection() bool {
	return false
}

const (
	hdrContentType           = "Content-Type"
	textType                 = "text/plain"
	textVersion              = "0.0.4"
	openmetricsType          = "application/openmetrics-text"
	openmetricsVersion_1_0_0 = "1.0.0"
	openmetricsVersion_0_0_1 = "0.0.1"
	protoType                = `application/vnd.google.protobuf`
	protoProtocol            = `io.prometheus.client.MetricFamily`
)

// responseFormat extracts the correct format from a HTTP response header.
// If no matching format can be found FormatUnknown is returned.
func responseFormat(h http.Header) expfmt.Format {
	ct := h.Get(hdrContentType)

	mediatype, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return expfmt.FmtUnknown
	}

	version := params["version"]

	switch mediatype {
	case protoType:
		p := params["proto"]
		e := params["encoding"]
		// only delimited encoding is supported by prometheus scraper
		if p == protoProtocol && e == "delimited" {
			return expfmt.FmtProtoDelim
		}

	case textType:
		if version == textVersion {
			return expfmt.FmtText
		}

	case openmetricsType:
		if version == openmetricsVersion_0_0_1 {
			return expfmt.FmtOpenMetrics_0_0_1
		}
		if version == openmetricsVersion_1_0_0 {
			return expfmt.FmtOpenMetrics_1_0_0
		}
	}

	return expfmt.FmtUnknown
}
