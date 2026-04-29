package envoy

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
	"github.com/kumahq/kuma/v2/pkg/xds/envoy/tags"
)

var _ = Describe("Metadata()", func() {
	It("should handle `nil` map of tags", func() {
		// when
		metadata := EndpointMetadata(nil)
		// then
		Expect(metadata).To(BeNil())
	})

	It("should handle empty map of tags", func() {
		// when
		metadata := EndpointMetadata(map[string]string{})
		// then
		Expect(metadata).To(BeNil())
	})

	It("should skip service tag", func() {
		// when
		metadata := EndpointMetadata(map[string]string{
			"kuma.io/service": "backend",
		})
		// then
		Expect(metadata).To(BeNil())
	})

	type testCase struct {
		tags     map[string]string
		expected string
	}
	DescribeTable("should generate Envoy metadata", func(given testCase) {
		// when
		metadata := EndpointMetadata(given.tags)
		// and
		actual, err := util_proto.ToYAML(metadata)
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(given.expected))
	},
		Entry("map with multiple tags", testCase{
			tags: map[string]string{
				"kuma.io/service": "redis",
				"version":         "v1",
				"region":          "eu",
			},
			expected: `
              filterMetadata:
                envoy.lb:
                  region: eu
                  version: v1
                envoy.transport_socket_match:
                  region: eu
                  version: v1`,
		}),
	)
})

var _ = Describe("EndpointMetadataWithLabels()", func() {
	It("should return tag-based metadata when tags are present, ignoring labels", func() {
		// given: tags produce non-nil metadata, so labels must not be added
		t := tags.Tags{
			"version": "v2",
			"region":  "us",
		}
		labels := map[string]string{"app": "backend"}

		// when
		metadata := EndpointMetadataWithLabels(t, labels)

		// then: metadata comes from tags, not labels
		Expect(metadata).ToNot(BeNil())
		Expect(metadata.GetFilterMetadata()).To(HaveKey("envoy.lb"))
		Expect(metadata.GetFilterMetadata()).ToNot(HaveKey(LbLabelsKey))
	})

	It("should return label-based metadata under LbLabelsKey when tags are absent", func() {
		// given: nil tags simulate KUMA_EXPERIMENTAL_INBOUND_TAGS_DISABLED
		labels := map[string]string{
			"app":     "frontend",
			"version": "v3",
		}

		// when
		metadata := EndpointMetadataWithLabels(nil, labels)

		// then: labels are placed under io.kuma.labels
		Expect(metadata).ToNot(BeNil())
		Expect(metadata.GetFilterMetadata()).To(HaveKey(LbLabelsKey))
		Expect(metadata.GetFilterMetadata()).ToNot(HaveKey("envoy.lb"))
		fields := metadata.GetFilterMetadata()[LbLabelsKey].GetFields()
		Expect(fields["app"].GetStringValue()).To(Equal("frontend"))
		Expect(fields["version"].GetStringValue()).To(Equal("v3"))
	})

	It("should return nil when tags are absent and labels are empty", func() {
		// when
		metadata := EndpointMetadataWithLabels(nil, map[string]string{})

		// then
		Expect(metadata).To(BeNil())
	})
})

var _ = Describe("ExtractLbLabels()", func() {
	It("should return empty tags for nil metadata", func() {
		// when
		result := ExtractLbLabels(nil)
		// then
		Expect(result).To(BeEmpty())
	})

	It("should return empty tags when LbLabelsKey is absent from metadata", func() {
		// given: regular tag-based metadata has no LbLabelsKey entry
		metadata := EndpointMetadata(tags.Tags{"region": "eu", "version": "v1"})

		// when
		result := ExtractLbLabels(metadata)

		// then
		Expect(result).To(BeEmpty())
	})

	It("should round-trip labels through EndpointMetadataWithLabels and ExtractLbLabels", func() {
		// given
		labels := map[string]string{
			"app":  "worker",
			"team": "platform",
		}

		// when
		metadata := EndpointMetadataWithLabels(nil, labels)
		result := ExtractLbLabels(metadata)

		// then
		Expect(result).To(HaveLen(2))
		Expect(result["app"]).To(Equal("worker"))
		Expect(result["team"]).To(Equal("platform"))
	})
})
