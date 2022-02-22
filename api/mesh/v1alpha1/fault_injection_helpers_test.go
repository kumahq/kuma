package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

var _ = Describe("FaultInjectionHelper", func() {

	Describe("SourceTags", func() {
		type testCase struct {
			input    *FaultInjection
			expected []SingleValueTagSet
		}
		DescribeTable("should return tags set for each selector",
			func(given testCase) {
				// when
				actual := given.input.SourceTags()
				// then
				Expect(actual).To(ConsistOf(given.expected))
			},
			Entry("basic", testCase{
				input: &FaultInjection{
					Sources: []*Selector{
						{
							Match: SingleValueTagSet{
								"tag1": "value1",
								"tag2": "value2",
							},
						},
						{
							Match: SingleValueTagSet{
								"tag3": "value3",
								"tag4": "value4",
							},
						},
					},
				},
				expected: []SingleValueTagSet{
					{
						"tag1": "value1",
						"tag2": "value2",
					},
					{
						"tag3": "value3",
						"tag4": "value4",
					},
				},
			}))
	})
})
