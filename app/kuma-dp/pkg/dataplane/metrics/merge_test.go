package metrics

import (
	"bufio"
	"bytes"
	"io"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
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
		Entry("should not merge unmergeable clusters for Counter with labels", testCase{
			input:    "./testdata/counter-unmergeable.in",
			expected: "./testdata/counter-unmergeable.out",
		}),
		Entry("should merge clusters for Counter", testCase{
			input:    "./testdata/counter-sparse.in",
			expected: "./testdata/counter-sparse.out",
		}),
	)
})
