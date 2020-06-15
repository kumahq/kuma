package generator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/xds/generator"
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
			input: []string{"a", "b", "c", "d", "e"},
			expected: [][]string{
				{"a"},
				{"b"},
				{"c"},
				{"d"},
				{"e"},
				{"a", "b"},
				{"a", "c"},
				{"a", "d"},
				{"a", "e"},
				{"b", "c"},
				{"b", "d"},
				{"b", "e"},
				{"c", "d"},
				{"c", "e"},
				{"d", "e"},
				{"a", "b", "c"},
				{"a", "b", "d"},
				{"a", "b", "e"},
				{"a", "c", "d"},
				{"a", "c", "e"},
				{"a", "d", "e"},
				{"b", "c", "d"},
				{"b", "c", "e"},
				{"b", "d", "e"},
				{"c", "d", "e"},
				{"a", "b", "c", "d"},
				{"a", "b", "c", "e"},
				{"a", "b", "d", "e"},
				{"a", "c", "d", "e"},
				{"b", "c", "d", "e"},
				{"a", "b", "c", "d", "e"},
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
