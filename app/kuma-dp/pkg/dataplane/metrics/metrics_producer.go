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

	"github.com/kumahq/kuma/v2/pkg/core"
)

var log = core.Log.WithName("metrics-custom-producer")

type AggregatedProducer struct {
	kumaVersion               string
	httpClientIPv4            http.Client
	httpClientIPv6            http.Client
	AppToScrape               ApplicationToScrape
	applicationsToScrape      []ApplicationToScrape
	applicationsToScrapeMutex *sync.Mutex
}

var _ sdkmetric.Producer = &AggregatedProducer{}

func NewAggregatedMetricsProducer(applicationsToScrape []ApplicationToScrape, isUsingTransparentProxy bool, kumaVersion string) *AggregatedProducer {
	return &AggregatedProducer{
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

	if len(appsToScrape) > 0 {
		appNames := make([]string, 0, len(appsToScrape))
		for _, app := range appsToScrape {
			appNames = append(appNames, app.Name)
		}
		log.V(1).Info("starting metrics collection", "applications", appNames, "count", len(appsToScrape))
	}

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
		result := combineMetrics(out)
		if len(result) == 0 && len(appsToScrape) > 0 {
			log.Info("metrics collection completed with zero metrics", "applicationCount", len(appsToScrape))
		}
		return result, nil
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
	if resp.StatusCode != http.StatusOK {
		log.Info("application returned non-200 status", "name", app.Name, "status", resp.StatusCode, "path", app.Path, "port", app.Port)
		return nil
	}
	requestTime := time.Now().UTC()

	metricsFromApplication, err := app.MeshMetricMutator(resp.Body)
	if err != nil {
		log.Error(err, "failed to mutate metrics")
		return nil
	}
	if len(metricsFromApplication) == 0 {
		log.Info("application returned empty metrics after parsing", "name", app.Name, "path", app.Path, "port", app.Port)
	}
	return FromPrometheusMetrics(metricsFromApplication, ap.kumaVersion, app.ExtraAttributes, requestTime)
}

func (ap *AggregatedProducer) makeRequest(ctx context.Context, req *http.Request, isIPv6 bool) (*http.Response, error) {
	req = req.WithContext(ctx)
	if isIPv6 {
		return ap.httpClientIPv6.Do(req) // #nosec G704 -- internal metrics scraping, operator-configured targets
	} else {
		return ap.httpClientIPv4.Do(req) // #nosec G704 -- internal metrics scraping, operator-configured targets
	}
}
