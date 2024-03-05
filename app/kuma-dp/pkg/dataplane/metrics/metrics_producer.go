package metrics

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"sync"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"golang.org/x/exp/maps"

	"github.com/kumahq/kuma/pkg/core"
)

var log = core.Log.WithName("metrics-custom-producer")

type AggregatedProducer struct {
	mesh                      string
	dataplane                 string
	service                   string
	httpClientIPv4            http.Client
	httpClientIPv6            http.Client
	AppToScrape               ApplicationToScrape
	applicationsToScrape      []ApplicationToScrape
	applicationsToScrapeMutex *sync.Mutex
}

var _ sdkmetric.Producer = &AggregatedProducer{}

func NewAggregatedMetricsProducer(mesh string, dataplane string, service string, applicationsToScrape []ApplicationToScrape, isUsingTransparentProxy bool) *AggregatedProducer {
	return &AggregatedProducer{
		mesh:                      mesh,
		dataplane:                 dataplane,
		service:                   service,
		httpClientIPv4:            createHttpClient(isUsingTransparentProxy, inPassThroughIPv4),
		httpClientIPv6:            createHttpClient(isUsingTransparentProxy, inPassThroughIPv6),
		applicationsToScrape:      applicationsToScrape,
		applicationsToScrapeMutex: &sync.Mutex{},
	}
}

func (ap *AggregatedProducer) SetApplicationsToScrape(applicationsToScrape []ApplicationToScrape) {
	ap.applicationsToScrapeMutex.Lock()
	defer ap.applicationsToScrapeMutex.Unlock()
	ap.applicationsToScrape = applicationsToScrape
}

func (ap *AggregatedProducer) Produce(ctx context.Context) ([]metricdata.ScopeMetrics, error) {
	ap.applicationsToScrapeMutex.Lock()
	var appsToScrape []ApplicationToScrape
	appsToScrape = append(appsToScrape, ap.applicationsToScrape...)
	ap.applicationsToScrapeMutex.Unlock()

	out := make(chan *metricdata.ScopeMetrics, len(appsToScrape))
	var wg sync.WaitGroup
	done := make(chan []byte)
	wg.Add(len(appsToScrape))
	go func() {
		wg.Wait()
		close(out)
		close(done)
	}()

	for _, app := range appsToScrape {
		go func(app ApplicationToScrape) {
			defer wg.Done()
			content := ap.fetchStats(ctx, app)
			out <- content
		}(app)
	}

	select {
	case <-ctx.Done():
		return nil, nil
	case <-done:
		return combineMetrics(out), nil
	}
}

func combineMetrics(metrics <-chan *metricdata.ScopeMetrics) []metricdata.ScopeMetrics {
	var combinedMetrics []metricdata.ScopeMetrics
	for metric := range metrics {
		if metric != nil {
			combinedMetrics = append(combinedMetrics, *metric)
		}
	}
	return combinedMetrics
}

func (ap *AggregatedProducer) fetchStats(ctx context.Context, app ApplicationToScrape) *metricdata.ScopeMetrics {
	req, err := http.NewRequest("GET", rewriteMetricsURL(app.Address, app.Port, app.Path, app.QueryModifier, &url.URL{}), nil)
	if err != nil {
		log.Error(err, "failed to create request")
		return nil
	}
	resp, err := ap.makeRequest(ctx, req, app.IsIPv6)
	if err != nil {
		log.Error(err, "failed call", "name", app.Name, "path", app.Path, "port", app.Port)
		return nil
	}
	defer resp.Body.Close()

	metricsFromApplication, err := app.OtelMutator(resp.Body)
	if err != nil {
		log.Error(err, "failed to mutate metrics")
		return nil
	}
	return &metricdata.ScopeMetrics{
		Scope: instrumentation.Scope{
			Name: app.Name,
		},
		Metrics: FromPrometheusMetrics(metricsFromApplication, ap.mesh, ap.dataplane, ap.service),
	}
}

func (ap *AggregatedProducer) makeRequest(ctx context.Context, req *http.Request, isIPv6 bool) (*http.Response, error) {
	req = req.WithContext(ctx)
	if isIPv6 {
		return ap.httpClientIPv6.Do(req)
	} else {
		return ap.httpClientIPv4.Do(req)
	}
}

func ParsePrometheusMetrics(in io.Reader) ([]*io_prometheus_client.MetricFamily, error) {
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(in)
	if err != nil {
		return nil, err
	}
	return maps.Values(metricFamilies), nil
}
