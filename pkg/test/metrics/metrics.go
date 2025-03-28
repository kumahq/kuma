package metrics

import (
	"github.com/onsi/gomega"
	prometheus_client "github.com/prometheus/client_model/go"

	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

func FindMetric(metrics core_metrics.Metrics, name string, labelsValues ...string) *prometheus_client.Metric {
	gathered, err := metrics.Gather()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	for _, metricFamily := range gathered {
		if metricFamily.Name != nil && *metricFamily.Name == name {
			for _, metric := range metricFamily.Metric {
				metricLabels := mapOfLabels(metric.Label)
				found := true
				for i := 0; i < len(labelsValues); i += 2 {
					key := labelsValues[i]
					value := labelsValues[i+1]
					found = found && metricLabels[key] == value
				}
				if found {
					return metric
				}
			}
		}
	}
	return nil
}

func mapOfLabels(pairs []*prometheus_client.LabelPair) map[string]string {
	result := map[string]string{}
	for _, pair := range pairs {
		result[*pair.Name] = *pair.Value
	}
	return result
}
