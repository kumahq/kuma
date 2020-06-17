package envoy

import (
	util_proto "github.com/Kong/kuma/pkg/util/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metadata()", func() {

	It("should handle `nil` map of tags", func() {
		// when
		metadata := Metadata(nil)
		// then
		Expect(metadata).To(BeNil())
	})

	It("should handle empty map of tags", func() {
		// when
		metadata := Metadata(map[string]string{})
		// then
		Expect(metadata).To(BeNil())
	})

	It("should skip service tag", func() {
		// when
		metadata := Metadata(map[string]string{
			"service": "backend",
		})
		// then
		Expect(metadata).To(BeNil())
	})

	type testCase struct {
		tags     map[string]string
		expected string
	}
	DescribeTable("should generate Envoy metadata",
		func(given testCase) {
			// when
			metadata := Metadata(given.tags)
			// and
			actual, err := util_proto.ToYAML(metadata)
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("map with multiple tags", testCase{
			tags: map[string]string{
				"service": "redis",
				"version": "v1",
				"region":  "eu",
			},
			expected: `
                filterMetadata:
                  envoy.lb:
                    version: v1
                    region: eu
`,
		}),
	)
})
