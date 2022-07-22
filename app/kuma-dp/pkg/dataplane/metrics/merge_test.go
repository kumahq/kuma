package metrics

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

func toLines(r io.Reader) (lines []string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return
}

var _ = Describe("Merge", func() {
	type testCase struct {
		input    string
		expected string
	}
	DescribeTable("should merge clusters in metrics",
		func(given testCase) {
			expected, err := os.Open(path.Join("testdata", given.expected))
			Expect(err).ToNot(HaveOccurred())
			input, err := os.Open(path.Join("testdata", given.input))
			Expect(err).ToNot(HaveOccurred())

			actual := new(bytes.Buffer)
			err = MergeClusters(input, actual)
			Expect(err).ToNot(HaveOccurred())

			Expect(toLines(actual)).To(ConsistOf(toLines(expected)))
		},
		Entry("should merge clusters for Counter", testCase{
			input:    "counter.in",
			expected: "counter.out",
		}),
		Entry("should merge clusters for Gauge", testCase{
			input:    "gauge.in",
			expected: "gauge.out",
		}),
		Entry("should merge clusters for Histogram", testCase{
			input:    "histogram.in",
			expected: "histogram.out",
		}),
		Entry("should merge clusters for Summary", testCase{
			input:    "summary.in",
			expected: "summary.out",
		}),
		Entry("should merge clusters for Untyped", testCase{
			input:    "untyped.in",
			expected: "untyped.out",
		}),
		Entry("should merge clusters for Counter with non-cluster metrics", testCase{
			input:    "counter-and-noncluster-metrics.in",
			expected: "counter-and-noncluster-metrics.out",
		}),
		Entry("should merge clusters for Counter with labels", testCase{
			input:    "counter-with-labels.in",
			expected: "counter-with-labels.out",
		}),
		Entry("should merge clusters for sparse Counter", testCase{
			input:    "counter-sparse.in",
			expected: "counter-sparse.out",
		}),
		Entry("should merge clusters for Counter with status codes label", testCase{
			input:    "counter-status-codes.in",
			expected: "counter-status-codes.out",
		}),
	)
})

var _ = Describe("Detect mergable clusters", func() {
	It("should crack split cluster names", func() {
		clusterName := envoy_names.GetSplitClusterName("foo-service", 99)
		Expect(clusterName).ToNot(BeEmpty())

		name, ok := isMergeableClusterName(clusterName)
		Expect(ok).To(BeTrue())
		Expect(name).To(Equal("foo-service"))
	})

	It("should crack gateway cluster names", func() {
		clusterName, err := route.DestinationClusterName(
			&route.Destination{
				Destination: map[string]string{
					mesh_proto.ServiceTag: "foo-service",
					mesh_proto.ZoneTag:    "foo-zone",
				},
			},
			map[string]string{
				"custom": "tag",
			},
		)
		Expect(err).To(Succeed())
		Expect(clusterName).ToNot(BeEmpty())

		name, ok := isMergeableClusterName(clusterName)
		Expect(ok).To(BeTrue())
		Expect(name).To(Equal("foo-service"))
	})

	It("should ignore other names", func() {
		name, ok := isMergeableClusterName("foo-service")
		Expect(ok).To(BeFalse())
		Expect(name).To(BeEmpty())
	})
})
