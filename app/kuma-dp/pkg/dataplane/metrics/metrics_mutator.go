package metrics

import (
	"io"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type (
	MetricsMutator    func(in io.Reader, out io.Writer) error
	PrometheusMutator func(in map[string]*io_prometheus_client.MetricFamily) error
	MeshMetricMutator func(in io.Reader) (map[string]*io_prometheus_client.MetricFamily, error)
)

func AggregatedOtelMutator(metricsMutators ...PrometheusMutator) MeshMetricMutator {
	return func(in io.Reader) (map[string]*io_prometheus_client.MetricFamily, error) {
		var parser expfmt.TextParser
		metricFamilies, err := parser.TextToMetricFamilies(in)
		if err != nil {
			return nil, err
		}

		for _, m := range metricsMutators {
			err := m(metricFamilies)
			if err != nil {
				return nil, err
			}
		}

		return metricFamilies, nil
	}
}

func AggregatedMetricsMutator(metricsMutators ...PrometheusMutator) MetricsMutator {
	return func(in io.Reader, out io.Writer) error {
		var parser expfmt.TextParser
		metricFamilies, err := parser.TextToMetricFamilies(in)
		if err != nil {
			return err
		}

		for _, m := range metricsMutators {
			err := m(metricFamilies)
			if err != nil {
				return err
			}
		}

		for _, metricFamily := range metricFamilies {
			if _, err := expfmt.MetricFamilyToText(out, metricFamily); err != nil {
				return err
			}
			if _, err := out.Write([]byte("\n")); err != nil {
				return err
			}
		}

		return nil
	}
}
