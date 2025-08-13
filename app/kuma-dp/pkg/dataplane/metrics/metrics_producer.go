package metrics

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/kumahq/kuma/pkg/core"
)

var log = core.Log.WithName("metrics-custom-producer")

type AggregatedProducer struct {
	mesh                      string
	dataplane                 string
	service                   string
	kumaVersion               string
	httpClientIPv4            http.Client
	httpClientIPv6            http.Client
	AppToScrape               ApplicationToScrape
	applicationsToScrape      []ApplicationToScrape
	applicationsToScrapeMutex *sync.Mutex
}

var _ sdkmetric.Producer = &AggregatedProducer{}

func NewAggregatedMetricsProducer(mesh string, dataplane string, service string, applicationsToScrape []ApplicationToScrape, isUsingTransparentProxy bool, kumaVersion string) *AggregatedProducer {
	return &AggregatedProducer{
		mesh:                      mesh,
		dataplane:                 dataplane,
		service:                   service,
		kumaVersion:               kumaVersion,
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

	out := make(chan map[instrumentation.Scope][]metricdata.Metrics, len(appsToScrape))
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

func combineMetrics(metricsChan <-chan map[instrumentation.Scope][]metricdata.Metrics) []metricdata.ScopeMetrics {
	aggregatedMetrics := map[instrumentation.Scope][]metricdata.Metrics{}
	for scopedMetrics := range metricsChan {
		for scope, metrics := range scopedMetrics {
			if _, ok := aggregatedMetrics[scope]; !ok {
				aggregatedMetrics[scope] = []metricdata.Metrics{}
			}
			aggregatedMetrics[scope] = append(aggregatedMetrics[scope], metrics...)
		}
	}

	var combinedMetrics []metricdata.ScopeMetrics
	for scope, metrics := range aggregatedMetrics {
		combinedMetrics = append(combinedMetrics, metricdata.ScopeMetrics{
			Scope:   scope,
			Metrics: metrics,
		})
	}

	return combinedMetrics
}

func (ap *AggregatedProducer) fetchStats(ctx context.Context, app ApplicationToScrape) map[instrumentation.Scope][]metricdata.Metrics {
	req, err := http.NewRequest(http.MethodGet, rewriteMetricsURL(app.Address, app.Port, app.Path, app.QueryModifier, &url.URL{}), http.NoBody)
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
	requestTime := time.Now().UTC()

	metricsFromApplication, err := app.MeshMetricMutator(resp.Body)
	if err != nil {
		log.Error(err, "failed to mutate metrics")
		return nil
	}
	return FromPrometheusMetrics(metricsFromApplication, ap.mesh, ap.dataplane, ap.service, ap.kumaVersion, app.ExtraLabels, requestTime)
}

func (ap *AggregatedProducer) makeRequest(ctx context.Context, req *http.Request, isIPv6 bool) (*http.Response, error) {
	req = req.WithContext(ctx)
	if isIPv6 {
		return ap.httpClientIPv6.Do(req)
	} else {
		return ap.httpClientIPv4.Do(req)
	}
}
