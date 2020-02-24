package accesslog_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/envoy/accesslog"
)

var _ = Describe("FilterStateOperator", func() {

	Describe("String()", func() {
		type testCase struct {
			key       string
			maxLength int
			expected  string
		}

		DescribeTable("should return correct canonical representation",
			func(given testCase) {
				// setup
				fragment := &FilterStateOperator{Key: given.key, MaxLength: given.maxLength}

				// when
				actual := fragment.String()
				// then
				Expect(actual).To(Equal(given.expected))

			},
			Entry("%FILTER_STATE()%", testCase{
				expected: `%FILTER_STATE()%`,
			}),
			Entry("%FILTER_STATE(filter.state.key)%", testCase{
				key:      "filter.state.key",
				expected: `%FILTER_STATE(filter.state.key)%`,
			}),
			Entry("%FILTER_STATE(filter.state.key):10%", testCase{
				key:       "filter.state.key",
				maxLength: 10,
				expected:  `%FILTER_STATE(filter.state.key):10%`,
			}),
		)
	})
})
