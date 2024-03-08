package leastcommonversion_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/tools/ci/update-vulnerable-dependencies/leastcommonversion"
)

var _ = Describe("Least Common Version", func() {
	type testCase struct {
		current  string
		matrix   [][]string
		expected string
	}
	DescribeTable("Deduct",
		func(given testCase) {
			actual, err := leastcommonversion.Deduct(&leastcommonversion.Input{
				Current:       given.current,
				FixedVersions: given.matrix,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(Equal(given.expected))
		},
		Entry("no fix versions", testCase{
			current:  "1.0.0",
			matrix:   [][]string{},
			expected: "null",
		}),
		Entry("no need to update", testCase{
			current: "1.1.0",
			matrix: [][]string{
				{"1.0.3", "1.1.0"},
				{"1.0.4"},
				{"1.0.5"},
			},
			expected: "null",
		}),
		Entry("picks the lowest common version", testCase{
			current: "1.0.0",
			matrix: [][]string{
				{"1.0.3", "1.1.0"},
				{"1.0.5", "1.2.0"},
				{"1.0.2"},
			},
			expected: "1.0.5",
		}),
	)
})
