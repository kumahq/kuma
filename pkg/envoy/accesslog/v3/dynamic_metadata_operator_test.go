package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/envoy/accesslog/v3"
)

var _ = Describe("DynamicMetadataOperator", func() {

	Describe("String()", func() {
		type testCase struct {
			filterNamespace string
			path            []string
			maxLength       int
			expected        string
		}

		DescribeTable("should return correct canonical representation",
			func(given testCase) {
				// setup
				fragment := &DynamicMetadataOperator{FilterNamespace: given.filterNamespace, Path: given.path, MaxLength: given.maxLength}

				// when
				actual := fragment.String()
				// then
				Expect(actual).To(Equal(given.expected))

			},
			Entry("%DYNAMIC_METADATA()%", testCase{
				expected: `%DYNAMIC_METADATA()%`,
			}),
			Entry("%DYNAMIC_METADATA():10%", testCase{
				maxLength: 10,
				expected:  `%DYNAMIC_METADATA():10%`,
			}),
			Entry("%DYNAMIC_METADATA(com.test.my_filter)%", testCase{
				filterNamespace: "com.test.my_filter",
				expected:        `%DYNAMIC_METADATA(com.test.my_filter)%`,
			}),
			Entry("%DYNAMIC_METADATA(com.test.my_filter):10%", testCase{
				filterNamespace: "com.test.my_filter",
				maxLength:       10,
				expected:        `%DYNAMIC_METADATA(com.test.my_filter):10%`,
			}),
			Entry("%DYNAMIC_METADATA(com.test.my_filter:test_object)%", testCase{
				filterNamespace: "com.test.my_filter",
				path:            []string{"test_object"},
				expected:        `%DYNAMIC_METADATA(com.test.my_filter:test_object)%`,
			}),
			Entry("%DYNAMIC_METADATA(com.test.my_filter:test_object):10%", testCase{
				filterNamespace: "com.test.my_filter",
				path:            []string{"test_object"},
				maxLength:       10,
				expected:        `%DYNAMIC_METADATA(com.test.my_filter:test_object):10%`,
			}),
			Entry("%DYNAMIC_METADATA(com.test.my_filter:test_object:inner_key)%", testCase{
				filterNamespace: "com.test.my_filter",
				path:            []string{"test_object", "inner_key"},
				expected:        `%DYNAMIC_METADATA(com.test.my_filter:test_object:inner_key)%`,
			}),
			Entry("%DYNAMIC_METADATA(com.test.my_filter:test_object:inner_key):10%", testCase{
				filterNamespace: "com.test.my_filter",
				path:            []string{"test_object", "inner_key"},
				maxLength:       10,
				expected:        `%DYNAMIC_METADATA(com.test.my_filter:test_object:inner_key):10%`,
			}),
			Entry("%DYNAMIC_METADATA(:test_object:inner_key)%", testCase{
				path:     []string{"test_object", "inner_key"},
				expected: `%DYNAMIC_METADATA(:test_object:inner_key)%`,
			}),
			Entry("%DYNAMIC_METADATA(:test_object:inner_key):10%", testCase{
				path:      []string{"test_object", "inner_key"},
				maxLength: 10,
				expected:  `%DYNAMIC_METADATA(:test_object:inner_key):10%`,
			}),
		)
	})
})
