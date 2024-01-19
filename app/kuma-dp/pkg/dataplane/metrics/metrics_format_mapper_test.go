package metrics

import (
	"encoding/json"
	"github.com/kumahq/kuma/pkg/test/matchers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path"
	"sort"
	"strings"
)

func testCaseName(ginkgo FullGinkgoTInterface) string {
	nameSplit := strings.Split(ginkgo.Name(), " ")
	return nameSplit[len(nameSplit)-1]
}

var _ = Describe("", func() {
	DescribeTable("should convert from Prometheus metrics to OpenTelemetry backend", func() {
		// given
		name := testCaseName(GinkgoT())
		input, err := os.Open(path.Join("testdata", "otel", name+".in"))
		Expect(err).ToNot(HaveOccurred())

		// when
		metrics, err := ParsePrometheusMetrics(input)
		Expect(err).ToNot(HaveOccurred())
		openTelemetryMetrics := FromPrometheusMetrics(metrics, "default", "dpp-1", "test-service")

		// then
		sort.SliceStable(openTelemetryMetrics, func(i, j int) bool {
			return openTelemetryMetrics[i].Name < openTelemetryMetrics[j].Name
		})
		marshal, err := json.Marshal(openTelemetryMetrics)
		Expect(err).ToNot(HaveOccurred())
		Expect(marshal).To(matchers.MatchGoldenJSON(path.Join("testdata", "otel", name+".golden.json")))
	},
		Entry("counter"),
		Entry("gauge"),
		Entry("histogram"),
		Entry("summary"),
	)
})
