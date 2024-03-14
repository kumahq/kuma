package metrics

import (
    io_prometheus_client "github.com/prometheus/client_model/go"
    "github.com/prometheus/common/expfmt"
    "io"
)

type (
	MetricsMutator func(in io.Reader, out io.Writer) error
    PrometheusMutator func(in map[string]*io_prometheus_client.MetricFamily) error
)

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
