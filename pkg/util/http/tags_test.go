package http

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/onsi/gomega"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
)

var _ = ginkgo.Describe("Http Tags Util", func() {

	ginkgo.Describe("SerializeTags", func() {

		type testCase struct {
			input  v1alpha1.MultiValueTagSet
			expect string
		}

		DescribeTable("should serialize tags to string",
			func(given testCase) {
				// when
				actual := SerializeTags(given.input)
				// then
				gomega.Expect(actual).To(gomega.Equal(given.expect))
			},
			Entry("basic case", testCase{
				input: v1alpha1.MultiValueTagSet{
					"tag1": map[string]bool{
						"value1": true,
						"value2": true,
					},
					"tag2": map[string]bool{
						"value3": true,
						"value4": true,
					},
				},
				expect: "&tag1=value1,value2&tag2=value3,value4&",
			}),
			Entry("empty input", testCase{
				input:  v1alpha1.MultiValueTagSet{},
				expect: "",
			}),
		)
	})
})
