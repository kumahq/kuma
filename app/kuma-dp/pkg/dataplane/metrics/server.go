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
	"strings"
	"sync"
	"time"

	"github.com/bakito/go-log-logr-adapter/adapter"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	v1alpha12 "github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/plugin/v1alpha1"
)

var (
	inPassThroughIPv4 = &net.TCPAddr{IP: net.ParseIP("127.0.0.6")}
	inPassThroughIPv6 = &net.TCPAddr{IP: net.ParseIP("::6")}
)

var (
	prometheusRequestHeaders = []string{"accept", "user-agent", "x-prometheus-scrape-timeout-seconds"}
	logger                   = core.Log.WithName("metrics-hijacker")

	// holds prometheus content types in order of priority.
	prometheusPriorityContentType = []expfmt.Format{
		FmtOpenMetrics_1_0_0,
		FmtOpenMetrics_0_0_1,
		expfmt.NewFormat(expfmt.TypeTextPlain),
		expfmt.NewFormat(expfmt.TypeUnknown),
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

type QueryParametersModifier func(queryParameters url.Values) url.Values

func RemoveQueryParameters(_ url.Values) url.Values {
	return url.Values{}
}

func AddPrometheusFormat(queryParameters url.Values) url.Values {
	queryParameters.Add("format", "prometheus")
	queryParameters.Add("text_readouts", "")
	return queryParameters
}

func AddSidecarParameters(sidecar *v1alpha12.Sidecar) func(queryParameters url.Values) url.Values {
	values := v1alpha1.EnvoyMetricsFilter(sidecar)

	return func(queryParameters url.Values) url.Values {
		queryParameters.Set("usedonly", values.Get("usedonly"))
		return queryParameters
	}
}

func AggregatedQueryParametersModifier(modifiers ...QueryParametersModifier) QueryParametersModifier {
	return func(queryParameters url.Values) url.Values {
		q := queryParameters
		for _, m := range modifiers {
			q = m(q)
		}
		return q
	}
}

type ApplicationToScrape struct {
	Name              string
	Address           string
	Path              string
	Port              uint32
	IsIPv6            bool
	ExtraLabels       map[string]string
	QueryModifier     QueryParametersModifier
	Mutator           MetricsMutator
	MeshMetricMutator MeshMetricMutator
}

type Hijacker struct {
	socketPath           string
	httpClientIPv4       http.Client
	httpClientIPv6       http.Client
	applicationsToScrape []ApplicationToScrape
	producer             *AggregatedProducer
	prometheusHandler    http.Handler
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

func New(socketPath string, applicationsToScrape []ApplicationToScrape, isUsingTransparentProxy bool, producer *AggregatedProducer) *Hijacker {
	return &Hijacker{
		socketPath:           socketPath,
		httpClientIPv4:       createHttpClient(isUsingTransparentProxy, inPassThroughIPv4),
		httpClientIPv6:       createHttpClient(isUsingTransparentProxy, inPassThroughIPv6),
		applicationsToScrape: applicationsToScrape,
		producer:             producer,
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
		ErrorLog:          adapter.ToStd(logger),
	}

	promExporter, err := prometheus.New(prometheus.WithProducer(s.producer), prometheus.WithoutCounterSuffixes())
	if err != nil {
		return err
	}
	sdkmetric.NewMeterProvider(sdkmetric.WithReader(promExporter))
	s.prometheusHandler = promhttp.Handler()

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
		RawQuery: queryModifier(in.Query()).Encode(),
	}
	return u.String()
}

func (s *Hijacker) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if strings.HasPrefix(req.URL.Path, v1alpha1.PrometheusDataplaneStatsPath) {
		s.prometheusHandler.ServeHTTP(writer, req)
		return
	}

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
		selectedCt := selectContentType(contentTypes, req.Header)
		writer.Header().Set(hdrContentType, string(selectedCt))

		// aggregate metrics of target applications and attempt to make them
		// compatible with FmtOpenMetrics if it is the selected content type.
		metrics := processMetrics(out, selectedCt)
		if _, err := writer.Write(metrics); err != nil {
			logger.Error(err, "error while writing the response")
		}
	}
}

func processMetrics(contents <-chan []byte, contentType expfmt.Format) []byte {
	buf := new(bytes.Buffer)

	for metrics := range contents {
		// remove the EOF marker from the metrics, because we are
		// merging multiple metrics into one response.
		metrics = bytes.ReplaceAll(metrics, []byte("# EOF"), []byte(""))

		buf.Write(metrics)
		buf.Write([]byte("\n"))
	}

	processedMetrics := append(processNewlineChars(buf.Bytes()), '\n')
	buf.Reset()
	buf.Write(processedMetrics)

	if contentType == FmtOpenMetrics_1_0_0 || contentType == FmtOpenMetrics_0_0_1 {
		// make metrics OpenMetrics compliant
		buf.Write([]byte("# EOF\n"))
	}

	return buf.Bytes()
}

// processNewlineChars takes byte data and returns a new byte slice
// after trimming and deduplicating the newline characters.
func processNewlineChars(input []byte) []byte {
	var deduped []byte

	if len(input) == 0 {
		return nil
	}
	last := input[0]

	for i := 1; i < len(input); i++ {
		if last == '\n' && input[i] == last {
			continue
		}
		deduped = append(deduped, last)
		last = input[i]
	}
	deduped = append(deduped, last)

	return bytes.TrimSpace(deduped)
}

// selectContentType selects the highest priority content type supported by the applications.
// If no valid content type is returned by the applications, it negotiates content type based
// on Accept header of the scraper.
func selectContentType(contentTypes <-chan expfmt.Format, reqHeader http.Header) expfmt.Format {
	// Tracks highest negotiated content type priority.
	// Lower number means higher priority
	//
	// We can not simply use the highest priority content type i.e. `application/openmetrics-text`
	// and try to mutate the metrics to make it compatible with this type,
	// because:
	// - if the application is not supporting this type,
	//   custom metrics might not be compatible (more prone to failure).
	// - the user might be using older prom scraper.
	//
	// So it's better to choose the highest negotiated content type between the
	// target apps and the scraper.
	var ctPriority int32 = math.MaxInt32
	ct := expfmt.NewFormat(expfmt.TypeUnknown)
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

	// If no valid content type is returned by the target applications,
	// negotitate content type based on Accept header of the scraper.
	if ct == expfmt.NewFormat(expfmt.TypeUnknown) {
		ct = expfmt.Negotiate(reqHeader)
	}

	return ct
}

func (s *Hijacker) getStats(ctx context.Context, initReq *http.Request, app ApplicationToScrape) ([]byte, expfmt.Format) {
	req, err := http.NewRequest("GET", rewriteMetricsURL(app.Address, app.Port, app.Path, app.QueryModifier, initReq.URL), http.NoBody)
	if err != nil {
		logger.Error(err, "failed to create request")
		return nil, ""
	}
	s.passRequestHeaders(req.Header, initReq.Header)
	req = req.WithContext(ctx)
	var resp *http.Response
	logger.V(1).Info("executing get stats request", "address", app.Address, "port", app.Port, "path", app.Path)
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
	hdrContentType = "Content-Type"
	textType       = "text/plain"
)

// responseFormat extracts the correct format from a HTTP response header.
// If no matching format can be found FormatUnknown is returned.
func responseFormat(h http.Header) expfmt.Format {
	ct := h.Get(hdrContentType)

	mediatype, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return expfmt.NewFormat(expfmt.TypeUnknown)
	}

	version := params["version"]

	switch mediatype {
	case expfmt.ProtoType:
		p := params["proto"]
		e := params["encoding"]
		// only delimited encoding is supported by prometheus scraper
		if p == expfmt.ProtoProtocol && e == "delimited" {
			return expfmt.NewFormat(expfmt.TypeProtoDelim)
		}

	// if mediatype is `text/plain`, return Prometheus text format
	// without checking the version, as there are few exporters
	// which don't set the version param in the content-type header. ex: Envoy
	case textType:
		return expfmt.NewFormat(expfmt.TypeTextPlain)

	// if mediatype is OpenMetricsType, return expfmt.NewFormat(expfmt.TypeUnknown) for any version
	// other than "0.0.1", "1.0.0" and "".
	case expfmt.OpenMetricsType:
		if version == expfmt.OpenMetricsVersion_0_0_1 || version == "" {
			return FmtOpenMetrics_0_0_1
		}
		if version == expfmt.OpenMetricsVersion_1_0_0 {
			return FmtOpenMetrics_1_0_0
		}
	}

	return expfmt.NewFormat(expfmt.TypeUnknown)
}
