package envoy

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
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

var _ = Describe("EndpointMetadata() label fallback", func() {
	It("should encode resource labels under envoy.lb when inbound tags are absent", func() {
		// given: nil tags simulate KUMA_EXPERIMENTAL_INBOUND_TAGS_DISABLED, so the
		// endpoint's load-balancing identity comes from resource labels instead.
		labels := map[string]string{
			"app":     "frontend",
			"version": "v3",
		}

		// when: callers fold labels into the same envoy.lb key
		metadata := EndpointMetadata(labels)

		// then
		Expect(metadata).ToNot(BeNil())
		Expect(metadata.GetFilterMetadata()).To(HaveKey("envoy.lb"))
		fields := metadata.GetFilterMetadata()["envoy.lb"].GetFields()
		Expect(fields["app"].GetStringValue()).To(Equal("frontend"))
		Expect(fields["version"].GetStringValue()).To(Equal("v3"))
	})

	It("should round-trip labels through envoy.lb via ExtractLbTags", func() {
		// given
		labels := map[string]string{
			"app":  "worker",
			"team": "platform",
		}

		// when
		result := ExtractLbTags(EndpointMetadata(labels))

		// then
		Expect(result).To(HaveLen(2))
		Expect(result["app"]).To(Equal("worker"))
		Expect(result["team"]).To(Equal("platform"))
	})
})
