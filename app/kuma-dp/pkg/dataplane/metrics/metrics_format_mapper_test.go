package metrics

import (
	"encoding/json"
	"os"
	"path"
	"sort"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("Metrics format mapper", func() {
	DescribeTable("should convert from Prometheus metrics to OpenTelemetry backend", func() {
		// given
		name := CurrentSpecReport().LeafNodeText
		input, err := os.Open(path.Join("testdata", "otel", name+".in"))
		Expect(err).ToNot(HaveOccurred())

		// when
		metrics, err := AggregatedOtelMutator()(input)
		Expect(err).ToNot(HaveOccurred())
		openTelemetryMetrics := FromPrometheusMetrics(metrics, "default", "dpp-1", "test-service", map[string]string{"extraLabel": "test"}, time.Date(2024, 1, 1, 1, 1, 1, 1, time.UTC))

		// then
		sort.SliceStable(openTelemetryMetrics, func(i, j int) bool {
			return openTelemetryMetrics[i].Name < openTelemetryMetrics[j].Name
		})
		marshal, err := json.MarshalIndent(openTelemetryMetrics, "", "  ")
		Expect(err).ToNot(HaveOccurred())
		Expect(marshal).To(matchers.MatchGoldenJSON(path.Join("testdata", "otel", name+".golden.json")))
	},
		Entry("counter"),
		Entry("gauge"),
		Entry("histogram"),
		Entry("summary"),
	)
})
