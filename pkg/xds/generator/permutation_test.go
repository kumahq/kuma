package generator_test

import (
	"github.com/Kong/kuma/pkg/xds/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Permutation", func() {

	type testCase struct {
		input    []string
		expected [][]string
	}
	DescribeTable("should generate permutation of strings",
		func(given testCase) {
			actual := generator.Permutation(given.input)
			Expect(actual).To(Equal(given.expected))
		},
		Entry("basic", testCase{
			input: []string{"a", "b", "c", "d"},
			expected: [][]string{
				{"a"}, {"b"}, {"c"}, {"d"},
				{"a", "b"}, {"a", "c"}, {"a", "d"}, {"b", "c"}, {"b", "d"}, {"c", "d"},
				{"a", "b", "c"}, {"a", "b", "d"}, {"a", "c", "d"}, {"b", "c", "d"},
				{"a", "b", "c", "d"},
			},
		}),
		Entry("empty input", testCase{
			input:    []string{},
			expected: [][]string{},
		}),
		Entry("one element input", testCase{
			input:    []string{"a"},
			expected: [][]string{{"a"}},
		}),
	)
})
