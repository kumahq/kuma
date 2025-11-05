package metrics

import (
	"encoding/json"
	"os"
	"path"
	"sort"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/kumahq/kuma/pkg/test/matchers"
)

type Scoped struct {
	Scope   instrumentation.Scope `json:"scope"`
	Metrics []metricdata.Metrics  `json:"metrics"`
}

var _ = Describe("Metrics format mapper", func() {
	DescribeTable("should convert from Prometheus metrics to OpenTelemetry backend", func(unifiedNaming bool) {
		// given
		name := CurrentSpecReport().LeafNodeText
		input, err := os.Open(path.Join("testdata", "otel", name+".in"))
		Expect(err).ToNot(HaveOccurred())

		// when
		metrics, err := AggregatedOtelMutator()(input)
		Expect(err).ToNot(HaveOccurred())
		openTelemetryMetrics := FromPrometheusMetrics(
			metrics,
			"default",
			"dpp-1",
			"test-service",
			"0.0.0",
			map[string]string{"extraLabel": "test"},
			unifiedNaming,
			time.Date(2024, 1, 1, 1, 1, 1, 1, time.UTC),
		)

		// then
		marshal, err := json.MarshalIndent(flatten(openTelemetryMetrics), "", "  ")
		Expect(err).ToNot(HaveOccurred())
		Expect(marshal).To(matchers.MatchGoldenJSON(path.Join("testdata", "otel", name+".golden.json")))
	},
		Entry("counter", false),
		Entry("gauge", false),
		Entry("histogram", false),
		Entry("summary", false),
		Entry("counter_unified_naming", true),
		Entry("gauge_unified_naming", true),
		Entry("histogram_unified_naming", true),
		Entry("summary_unified_naming", true),
	)
})

func flatten(scoped map[instrumentation.Scope][]metricdata.Metrics) []Scoped {
	out := make([]Scoped, 0, len(scoped))
	for s, ms := range scoped {
		sort.SliceStable(ms, func(i, j int) bool {
			return ms[i].Name < ms[j].Name
		})
		out = append(out, Scoped{Scope: s, Metrics: ms})
	}

	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Scope.Name < out[j].Scope.Name
	})
	return out
}
