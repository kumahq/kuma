package metrics

import (
	"math"
	"strings"
	"time"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/kumahq/kuma/pkg/util/pointer"
)

const (
	otelScopePrefix    = "otel_scope_"
	otelScopeName      = "otel_scope_name"
	otelScopeVersion   = "otel_scope_version"
	otelScopeSchemaUrl = "otel_scope_schema_url"
	kumaOtelScope      = "kuma"
)

func FromPrometheusMetrics(appMetrics map[string]*io_prometheus_client.MetricFamily, mesh, dataplane, service, kumaVersion string, extraLabels map[string]string, requestTime time.Time) map[instrumentation.Scope][]metricdata.Metrics {
	extraAttributes := extraAttributesFrom(mesh, dataplane, service, extraLabels)

	scopedMetrics := map[instrumentation.Scope][]metricdata.Metrics{}
	for _, prometheusMetric := range appMetrics {
		var scopedAggregations map[instrumentation.Scope]metricdata.Aggregation
		switch prometheusMetric.GetType() {
		case io_prometheus_client.MetricType_GAUGE:
			scopedAggregations = scopedGauges(prometheusMetric.Metric, kumaVersion, extraAttributes, requestTime)
		case io_prometheus_client.MetricType_SUMMARY:
			scopedAggregations = scopedSummaries(prometheusMetric.Metric, kumaVersion, extraAttributes, requestTime)
		case io_prometheus_client.MetricType_COUNTER:
			scopedAggregations = scopedCounters(prometheusMetric.Metric, kumaVersion, extraAttributes, requestTime)
		case io_prometheus_client.MetricType_HISTOGRAM:
			scopedAggregations = scopedHistograms(prometheusMetric.Metric, kumaVersion, extraAttributes, requestTime)
		default:
			log.Info("got unsupported metric type", "type", prometheusMetric.Type)
		}

		for scope, aggregations := range scopedAggregations {
			scopedMetrics[scope] = append(scopedMetrics[scope], metricdata.Metrics{
				Name:        prometheusMetric.GetName(),
				Description: prometheusMetric.GetHelp(),
				Data:        aggregations,
			})
		}
	}

	return scopedMetrics
}

func scopedGauges(prometheusData []*io_prometheus_client.Metric, kumaVersion string, extraAttributes []attribute.KeyValue, requestTime time.Time) map[instrumentation.Scope]metricdata.Aggregation {
	scopedDataPoints := map[instrumentation.Scope][]metricdata.DataPoint[float64]{}
	for _, metric := range prometheusData {
		scope, attributes := extractScope(metric, kumaVersion)
		attributes = append(attributes, extraAttributes...)
		scopedDataPoints[scope] = append(scopedDataPoints[scope], metricdata.DataPoint[float64]{
			Attributes: attribute.NewSet(attributes...),
			Time:       getTimeOrFallback(metric.TimestampMs, requestTime),
			Value:      metric.Gauge.GetValue(),
		})
	}

	scopedAggregations := map[instrumentation.Scope]metricdata.Aggregation{}
	for scope, data := range scopedDataPoints {
		scopedAggregations[scope] = metricdata.Gauge[float64]{
			DataPoints: data,
		}
	}

	return scopedAggregations
}

func scopedSummaries(prometheusData []*io_prometheus_client.Metric, kumaVersion string, extraAttributes []attribute.KeyValue, requestTime time.Time) map[instrumentation.Scope]metricdata.Aggregation {
	scopedDataPoints := map[instrumentation.Scope][]metricdata.SummaryDataPoint{}
	for _, metric := range prometheusData {
		scope, attributes := extractScope(metric, kumaVersion)
		attributes = append(attributes, extraAttributes...)
		scopedDataPoints[scope] = append(scopedDataPoints[scope], metricdata.SummaryDataPoint{
			Attributes:     attribute.NewSet(attributes...),
			Time:           getTimeOrFallback(metric.TimestampMs, requestTime),
			QuantileValues: toOpenTelemetryQuantile(metric.Summary.Quantile),
			Sum:            pointer.Deref(metric.Summary.SampleSum),
			Count:          pointer.Deref(metric.Summary.SampleCount),
		})
	}

	scopedAggregations := map[instrumentation.Scope]metricdata.Aggregation{}
	for scope, data := range scopedDataPoints {
		scopedAggregations[scope] = metricdata.Summary{
			DataPoints: data,
		}
	}

	return scopedAggregations
}

func scopedCounters(prometheusData []*io_prometheus_client.Metric, kumaVersion string, extraAttributes []attribute.KeyValue, requestTime time.Time) map[instrumentation.Scope]metricdata.Aggregation {
	scopedDataPoints := map[instrumentation.Scope][]metricdata.DataPoint[float64]{}
	for _, metric := range prometheusData {
		scope, attributes := extractScope(metric, kumaVersion)
		attributes = append(attributes, extraAttributes...)
		scopedDataPoints[scope] = append(scopedDataPoints[scope], metricdata.DataPoint[float64]{
			Attributes: attribute.NewSet(attributes...),
			Time:       getTimeOrFallback(metric.TimestampMs, requestTime),
			Value:      metric.Counter.GetValue(),
		})
	}

	scopedAggregations := map[instrumentation.Scope]metricdata.Aggregation{}
	for scope, data := range scopedDataPoints {
		scopedAggregations[scope] = metricdata.Sum[float64]{
			IsMonotonic: true,
			Temporality: metricdata.CumulativeTemporality,
			DataPoints:  data,
		}
	}

	return scopedAggregations
}

func scopedHistograms(prometheusData []*io_prometheus_client.Metric, kumaVersion string, extraAttributes []attribute.KeyValue, requestTime time.Time) map[instrumentation.Scope]metricdata.Aggregation {
	scopedDataPoints := map[instrumentation.Scope][]metricdata.HistogramDataPoint[float64]{}
	for _, metric := range prometheusData {
		scope, attributes := extractScope(metric, kumaVersion)
		attributes = append(attributes, extraAttributes...)

		var bounds []float64
		var bucketCounts []uint64
		for idx, bucket := range metric.Histogram.Bucket {
			if !math.IsInf(bucket.GetUpperBound(), 1) {
				bounds = append(bounds, bucket.GetUpperBound())
			}

			bucketValue := bucket.GetCumulativeCount()
			// we need to get actual count from specific bucket, not from all smaller buckets to convert prometheus histogram to native histogram
			if idx > 0 {
				bucketValue -= metric.Histogram.Bucket[idx-1].GetCumulativeCount()
			}
			bucketCounts = append(bucketCounts, bucketValue)
		}

		scopedDataPoints[scope] = append(scopedDataPoints[scope], metricdata.HistogramDataPoint[float64]{
			Time:         getTimeOrFallback(metric.TimestampMs, requestTime),
			Attributes:   attribute.NewSet(attributes...),
			Count:        metric.Histogram.GetSampleCount(),
			Sum:          metric.Histogram.GetSampleSum(),
			Bounds:       bounds,
			BucketCounts: bucketCounts,
		})
	}

	scopedAggregations := map[instrumentation.Scope]metricdata.Aggregation{}
	for scope, data := range scopedDataPoints {
		scopedAggregations[scope] = metricdata.Histogram[float64]{
			Temporality: metricdata.CumulativeTemporality,
			DataPoints:  data,
		}
	}

	return scopedAggregations
}

func getTimeOrFallback(timestampMs *int64, fallback time.Time) time.Time {
	if timestampMs != nil {
		return time.UnixMilli(*timestampMs).UTC()
	} else {
		return fallback
	}
}

func toOpenTelemetryQuantile(prometheusQuantiles []*io_prometheus_client.Quantile) []metricdata.QuantileValue {
	var otelQuantiles []metricdata.QuantileValue
	for _, quantile := range prometheusQuantiles {
		otelQuantiles = append(otelQuantiles, metricdata.QuantileValue{
			Quantile: quantile.GetQuantile(),
			Value:    quantile.GetValue(),
		})
	}
	return otelQuantiles
}

func extraAttributesFrom(mesh, dataplane, service string, extraLabels map[string]string) []attribute.KeyValue {
	var extraAttributes []attribute.KeyValue
	if mesh != "" {
		extraAttributes = append(extraAttributes, attribute.String("mesh", mesh))
	}
	if dataplane != "" {
		extraAttributes = append(extraAttributes, attribute.String("dataplane", dataplane))
	}
	if service != "" {
		extraAttributes = append(extraAttributes, attribute.String("service", service))
	}
	for k, v := range extraLabels {
		extraAttributes = append(extraAttributes, attribute.String(k, v))
	}
	return extraAttributes
}

func extractScope(metric *io_prometheus_client.Metric, kumaVersion string) (instrumentation.Scope, []attribute.KeyValue) {
	var attributes []attribute.KeyValue
	var scopeAttributes []attribute.KeyValue
	var scope instrumentation.Scope
	for _, label := range metric.Label {
		if !strings.HasPrefix(label.GetName(), otelScopePrefix) {
			attributes = append(attributes, attribute.String(label.GetName(), label.GetValue()))
			continue
		}

		switch label.GetName() {
		case otelScopeName:
			scope.Name = label.GetValue()
		case otelScopeVersion:
			scope.Version = label.GetValue()
		case otelScopeSchemaUrl:
			scope.SchemaURL = label.GetValue()
		default:
			scopeAttributes = append(scopeAttributes, attribute.String(strings.TrimPrefix(label.GetName(), otelScopePrefix), label.GetValue()))
		}
	}

	if len(scopeAttributes) > 0 {
		scope.Attributes = attribute.NewSet(scopeAttributes...)
	}

	// If metrics were not scoped, we need to create Kuma scope for it
	if len(attributes) == len(metric.Label) {
		scope.Name = kumaOtelScope
		scope.Version = kumaVersion
	}

	return scope, attributes
}
