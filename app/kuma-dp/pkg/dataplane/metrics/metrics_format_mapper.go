package metrics

import (
	"math"
	"time"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/kumahq/kuma/pkg/util/pointer"
)

func FromPrometheusMetrics(appMetrics map[string]*io_prometheus_client.MetricFamily, mesh string, dataplane string, service string, extraLabels map[string]string, requestTime time.Time) []metricdata.Metrics {
	extraAttributes := extraAttributesFrom(mesh, dataplane, service, extraLabels)

	var openTelemetryMetrics []metricdata.Metrics
	for _, prometheusMetric := range appMetrics {
		otelMetric := metricdata.Metrics{
			Name:        prometheusMetric.GetName(),
			Description: prometheusMetric.GetHelp(),
		}

		switch prometheusMetric.GetType() {
		case io_prometheus_client.MetricType_GAUGE:
			otelMetric.Data = metricdata.Gauge[float64]{
				DataPoints: gaugeDataPoints(prometheusMetric.Metric, extraAttributes, requestTime),
			}
		case io_prometheus_client.MetricType_SUMMARY:
			otelMetric.Data = metricdata.Summary{
				DataPoints: summaryDataPoints(prometheusMetric.Metric, extraAttributes, requestTime),
			}
		case io_prometheus_client.MetricType_COUNTER:
			otelMetric.Data = metricdata.Sum[float64]{
				IsMonotonic: true,
				Temporality: metricdata.CumulativeTemporality,
				DataPoints:  counterDataPoints(prometheusMetric.Metric, extraAttributes, requestTime),
			}
		case io_prometheus_client.MetricType_HISTOGRAM:
			otelMetric.Data = metricdata.Histogram[float64]{
				Temporality: metricdata.CumulativeTemporality,
				DataPoints:  histogramDataPoints(prometheusMetric.Metric, extraAttributes, requestTime),
			}
		default:
			log.Info("got unsupported metric type", "type", prometheusMetric.Type)
		}
		openTelemetryMetrics = append(openTelemetryMetrics, otelMetric)
	}

	return openTelemetryMetrics
}

func gaugeDataPoints(prometheusData []*io_prometheus_client.Metric, extraAttributes []attribute.KeyValue, requestTime time.Time) []metricdata.DataPoint[float64] {
	var dataPoints []metricdata.DataPoint[float64]
	for _, metric := range prometheusData {
		attributes := createOpenTelemetryAttributes(metric.Label, extraAttributes)
		dataPoints = append(dataPoints, metricdata.DataPoint[float64]{
			Attributes: attributes,
			Time:       getTimeOrFallback(metric.TimestampMs, requestTime),
			Value:      metric.Gauge.GetValue(),
		})
	}
	return dataPoints
}

func summaryDataPoints(prometheusData []*io_prometheus_client.Metric, extraAttributes []attribute.KeyValue, requestTime time.Time) []metricdata.SummaryDataPoint {
	var dataPoints []metricdata.SummaryDataPoint
	for _, metric := range prometheusData {
		attributes := createOpenTelemetryAttributes(metric.Label, extraAttributes)
		dataPoints = append(dataPoints, metricdata.SummaryDataPoint{
			Attributes:     attributes,
			Time:           getTimeOrFallback(metric.TimestampMs, requestTime),
			QuantileValues: toOpenTelemetryQuantile(metric.Summary.Quantile),
			Sum:            pointer.Deref(metric.Summary.SampleSum),
			Count:          pointer.Deref(metric.Summary.SampleCount),
		})
	}
	return dataPoints
}

func counterDataPoints(prometheusData []*io_prometheus_client.Metric, extraAttributes []attribute.KeyValue, requestTime time.Time) []metricdata.DataPoint[float64] {
	var dataPoints []metricdata.DataPoint[float64]
	for _, metric := range prometheusData {
		attributes := createOpenTelemetryAttributes(metric.Label, extraAttributes)
		dataPoints = append(dataPoints, metricdata.DataPoint[float64]{
			Attributes: attributes,
			Time:       getTimeOrFallback(metric.TimestampMs, requestTime),
			Value:      metric.Counter.GetValue(),
		})
	}
	return dataPoints
}

func histogramDataPoints(prometheusData []*io_prometheus_client.Metric, extraAttributes []attribute.KeyValue, requestTime time.Time) []metricdata.HistogramDataPoint[float64] {
	var dataPoints []metricdata.HistogramDataPoint[float64]
	for _, metric := range prometheusData {
		attributes := createOpenTelemetryAttributes(metric.Label, extraAttributes)

		var bounds []float64
		var bucketCounts []uint64
		for _, bucket := range metric.Histogram.Bucket {
			if !math.IsInf(bucket.GetUpperBound(), 1) {
				bounds = append(bounds, bucket.GetUpperBound())
			}
			bucketCounts = append(bucketCounts, bucket.GetCumulativeCount())
		}

		dataPoints = append(dataPoints, metricdata.HistogramDataPoint[float64]{
			Time:         getTimeOrFallback(metric.TimestampMs, requestTime),
			Attributes:   attributes,
			Count:        metric.Histogram.GetSampleCount(),
			Sum:          metric.Histogram.GetSampleSum(),
			Bounds:       bounds,
			BucketCounts: bucketCounts,
		})
	}
	return dataPoints
}

func getTimeOrFallback(timestampMs *int64, fallback time.Time) time.Time {
	if timestampMs != nil {
		return time.UnixMilli(*timestampMs).UTC()
	} else {
		return fallback
	}
}

func createOpenTelemetryAttributes(labels []*io_prometheus_client.LabelPair, extraAttributes []attribute.KeyValue) attribute.Set {
	var attributes []attribute.KeyValue
	for _, label := range labels {
		attributes = append(attributes, attribute.String(label.GetName(), label.GetValue()))
	}
	attributes = append(attributes, extraAttributes...)
	return attribute.NewSet(attributes...)
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

func extraAttributesFrom(mesh string, dataplane string, service string, extraLabels map[string]string) []attribute.KeyValue {
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
