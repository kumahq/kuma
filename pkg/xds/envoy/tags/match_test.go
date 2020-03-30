package tags_test

import (
	"fmt"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	. "github.com/Kong/kuma/pkg/xds/envoy/tags"
)

var _ = Describe("MatchingRegex", func() {
	type testCase struct {
		serviceTags v1alpha1.MultiValueTagSet
		matchTags   v1alpha1.SingleValueTagSet
		expected    bool
	}
	DescribeTable("should generate regex for matching service's tags",
		func(given testCase) {
			// when
			regexStr := MatchingRegex(given.matchTags)
			fmt.Println(regexStr)
			re, err := regexp.Compile(regexStr)
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			matched := re.MatchString(Serialize(given.serviceTags))
			fmt.Println(Serialize(given.serviceTags))

			// then
			Expect(matched).To(Equal(given.expected))
		},
		Entry("match without middle tag2", testCase{
			serviceTags: v1alpha1.MultiValueTagSet{
				"tag1": {"value1": true, "value2": true},
				"tag2": {"value2": true, "value3": true},
				"tag3": {"value3": true, "value4": true},
			},
			matchTags: v1alpha1.SingleValueTagSet{
				"tag1": "value1",
				"tag3": "value3",
			},
			expected: true,
		}),
		Entry("shouldn't match", testCase{
			serviceTags: v1alpha1.MultiValueTagSet{
				"tag1": {"value1": true, "value2": true},
				"tag2": {"value2": true, "value3": true},
				"tag3": {"value3": true, "value4": true},
			},
			matchTags: v1alpha1.SingleValueTagSet{
				"tag1": "value1",
				"tag3": "value5",
			},
			expected: false,
		}),
		Entry("shouldn't match value's prefix", testCase{
			serviceTags: v1alpha1.MultiValueTagSet{
				"tag1": {"value1": true, "value2": true},
			},
			matchTags: v1alpha1.SingleValueTagSet{
				"tag1": "val",
			},
			expected: false,
		}),
		Entry("should match asterisk tag", testCase{
			serviceTags: v1alpha1.MultiValueTagSet{
				"tag1": {"value1": true, "value2": true},
			},
			matchTags: v1alpha1.SingleValueTagSet{
				"tag1": "*",
			},
			expected: true,
		}),
		Entry("shouldn't match asterisk tag", testCase{
			serviceTags: v1alpha1.MultiValueTagSet{
				"tag1": {"value1": true, "value2": true},
			},
			matchTags: v1alpha1.SingleValueTagSet{
				"tag2": "*",
			},
			expected: false,
		}),
	)
})
