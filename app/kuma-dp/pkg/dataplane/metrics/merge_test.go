package metrics

import (
	"bufio"
	"bytes"
	"io"
	"os"

	envoy_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
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
			input, err := os.Open(given.input)
			Expect(err).ToNot(HaveOccurred())
			expected, err := os.Open(given.expected)
			Expect(err).ToNot(HaveOccurred())

			actual := new(bytes.Buffer)
			err = MergeClusters(input, actual)
			Expect(err).ToNot(HaveOccurred())
			Expect(toLines(actual)).To(ConsistOf(toLines(expected)))
		},
		Entry("should merge clusters for Counter", testCase{
			input:    "./testdata/counter.in",
			expected: "./testdata/counter.out",
		}),
		Entry("should merge clusters for Gauge", testCase{
			input:    "./testdata/gauge.in",
			expected: "./testdata/gauge.out",
		}),
		Entry("should merge clusters for Histogram", testCase{
			input:    "./testdata/histogram.in",
			expected: "./testdata/histogram.out",
		}),
		Entry("should merge clusters for Summary", testCase{
			input:    "./testdata/summary.in",
			expected: "./testdata/summary.out",
		}),
		Entry("should merge clusters for Untyped", testCase{
			input:    "./testdata/untyped.in",
			expected: "./testdata/untyped.out",
		}),
		Entry("should merge clusters for Counter with non-cluster metrics", testCase{
			input:    "./testdata/counter-and-noncluster-metrics.in",
			expected: "./testdata/counter-and-noncluster-metrics.out",
		}),
		Entry("should merge clusters for Counter with labels", testCase{
			input:    "./testdata/counter-with-labels.in",
			expected: "./testdata/counter-with-labels.out",
		}),
		Entry("should merge clusters for Counter", testCase{
			input:    "./testdata/counter-sparse.in",
			expected: "./testdata/counter-sparse.out",
		}),
		Entry("should merge clusters for Counter with status codes label", testCase{
			input:    "./testdata/counter-status-codes.in",
			expected: "./testdata/counter-status-codes.out",
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
			&envoy_cluster_v3.Cluster{},
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
