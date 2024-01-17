package metrics

import (
	"math"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func FromPrometheusMetrics(appMetrics []*io_prometheus_client.MetricFamily) []metricdata.Metrics {
	var openTelemetryMetrics []metricdata.Metrics
	for _, prometheusMetric := range appMetrics {
		otelMetric := metricdata.Metrics{
			Name:        prometheusMetric.GetName(),
			Description: prometheusMetric.GetHelp(),
		}

		switch prometheusMetric.GetType() {
		case io_prometheus_client.MetricType_GAUGE:
			otelMetric.Data = metricdata.Gauge[float64]{
				DataPoints: gaugeDataPoints(prometheusMetric.Metric),
			}
		case io_prometheus_client.MetricType_SUMMARY:
			otelMetric.Data = metricdata.Summary{
				DataPoints: summaryDataPoints(prometheusMetric.Metric),
			}
		case io_prometheus_client.MetricType_COUNTER:
			otelMetric.Data = metricdata.Sum[float64]{
				IsMonotonic: true,
				Temporality: metricdata.CumulativeTemporality,
				DataPoints:  counterDataPoints(prometheusMetric.Metric),
			}
		case io_prometheus_client.MetricType_HISTOGRAM:
			otelMetric.Data = metricdata.Histogram[float64]{
				Temporality: metricdata.CumulativeTemporality,
				DataPoints:  histogramDataPoints(prometheusMetric.Metric),
			}
		default:
			log.Info("Got unsupported metric type", "type", prometheusMetric.Type)
		}
		openTelemetryMetrics = append(openTelemetryMetrics, otelMetric)
	}

	return openTelemetryMetrics
}

func gaugeDataPoints(prometheusData []*io_prometheus_client.Metric) []metricdata.DataPoint[float64] {
	var dataPoints []metricdata.DataPoint[float64]
	for _, metric := range prometheusData {
		attributes := createOpenTelemetryAttributes(metric.Label)
		dataPoints = append(dataPoints, metricdata.DataPoint[float64]{
			Attributes: attributes,
			Value:      metric.Gauge.GetValue(),
		})
	}
	return dataPoints
}

func summaryDataPoints(prometheusData []*io_prometheus_client.Metric) []metricdata.SummaryDataPoint {
	var dataPoints []metricdata.SummaryDataPoint
	for _, metric := range prometheusData {
		attributes := createOpenTelemetryAttributes(metric.Label)
		dataPoints = append(dataPoints, metricdata.SummaryDataPoint{
			Attributes:     attributes,
			QuantileValues: toOpenTelemetryQuantile(metric.Summary.Quantile),
			Time:           metric.Summary.CreatedTimestamp.AsTime(),
		})
	}
	return dataPoints
}

func counterDataPoints(prometheusData []*io_prometheus_client.Metric) []metricdata.DataPoint[float64] {
	var dataPoints []metricdata.DataPoint[float64]
	for _, metric := range prometheusData {
		attributes := createOpenTelemetryAttributes(metric.Label)
		dataPoints = append(dataPoints, metricdata.DataPoint[float64]{
			Attributes: attributes,
			Value:      metric.Counter.GetValue(),
		})
	}
	return dataPoints
}

func histogramDataPoints(prometheusData []*io_prometheus_client.Metric) []metricdata.HistogramDataPoint[float64] {
	var dataPoints []metricdata.HistogramDataPoint[float64]
	for _, metric := range prometheusData {
		attributes := createOpenTelemetryAttributes(metric.Label)

		var bounds []float64
		var bucketCounts []uint64
		for _, bucket := range metric.Histogram.Bucket {
			if !math.IsInf(bucket.GetUpperBound(), 1) {
				bounds = append(bounds, bucket.GetUpperBound())
			}
			bucketCounts = append(bucketCounts, bucket.GetCumulativeCount())
		}

		dataPoints = append(dataPoints, metricdata.HistogramDataPoint[float64]{
			Attributes:   attributes,
			Count:        metric.Histogram.GetSampleCount(),
			Sum:          metric.Histogram.GetSampleSum(),
			Bounds:       bounds,
			BucketCounts: bucketCounts,
		})
	}
	return dataPoints
}

func createOpenTelemetryAttributes(labels []*io_prometheus_client.LabelPair) attribute.Set {
	var attributes []attribute.KeyValue
	for _, label := range labels {
		attributes = append(attributes, attribute.String(label.GetName(), label.GetValue()))
	}
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
