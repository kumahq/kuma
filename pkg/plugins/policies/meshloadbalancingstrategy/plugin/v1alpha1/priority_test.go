package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	v1alpha1 "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshloadbalancingstrategy/plugin/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

var _ = Describe("GetLocalityGroups()", func() {
	Describe("label fallback when inbound tags are disabled", func() {
		It("should build local lb groups from pod labels when inbound tags are absent", func() {
			// given: configuration with a single affinity tag
			conf := &api.Conf{
				LocalityAwareness: &api.LocalityAwareness{
					LocalZone: &api.LocalZone{
						AffinityTags: pointer.To([]api.AffinityTag{
							{Key: "region"},
						}),
					},
				},
			}

			// given: no inbound tags (simulates KUMA_EXPERIMENTAL_INBOUND_TAGS_DISABLED)
			var inboundTags mesh_proto.MultiValueTagSet

			// given: pod labels carry the affinity key
			podLabels := map[string]string{
				"region": "eu-west",
			}

			// when
			localGroups, _ := v1alpha1.GetLocalityGroups(conf, inboundTags, podLabels, "local-zone")

			// then: pod label value is used as the locality group value
			Expect(localGroups).To(HaveLen(1))
			Expect(localGroups[0].Key).To(Equal("region"))
			Expect(localGroups[0].Value).To(Equal("eu-west"))
		})

		It("should build local lb groups from pod labels for multiple affinity tags", func() {
			// given: configuration with two affinity tags
			weight1 := uint32(80)
			weight2 := uint32(20)
			conf := &api.Conf{
				LocalityAwareness: &api.LocalityAwareness{
					LocalZone: &api.LocalZone{
						AffinityTags: pointer.To([]api.AffinityTag{
							{Key: "region", Weight: &weight1},
							{Key: "zone", Weight: &weight2},
						}),
					},
				},
			}

			// given: no inbound tags
			var inboundTags mesh_proto.MultiValueTagSet

			// given: pod labels supply both affinity keys
			podLabels := map[string]string{
				"region": "us-east",
				"zone":   "us-east-1a",
			}

			// when
			localGroups, _ := v1alpha1.GetLocalityGroups(conf, inboundTags, podLabels, "local-zone")

			// then: both affinity tags are resolved from pod labels
			Expect(localGroups).To(HaveLen(2))
			Expect(localGroups[0].Key).To(Equal("region"))
			Expect(localGroups[0].Value).To(Equal("us-east"))
			Expect(localGroups[0].Weight).To(Equal(uint32(80)))
			Expect(localGroups[1].Key).To(Equal("zone"))
			Expect(localGroups[1].Value).To(Equal("us-east-1a"))
			Expect(localGroups[1].Weight).To(Equal(uint32(20)))
		})

		It("should produce no local lb groups when pod labels also lack the affinity key", func() {
			// given
			conf := &api.Conf{
				LocalityAwareness: &api.LocalityAwareness{
					LocalZone: &api.LocalZone{
						AffinityTags: pointer.To([]api.AffinityTag{
							{Key: "region"},
						}),
					},
				},
			}

			// given: neither inbound tags nor pod labels supply the key
			var inboundTags mesh_proto.MultiValueTagSet
			podLabels := map[string]string{}

			// when
			localGroups, _ := v1alpha1.GetLocalityGroups(conf, inboundTags, podLabels, "local-zone")

			// then: no groups are produced
			Expect(localGroups).To(BeEmpty())
		})

		It("should prefer inbound tags over pod labels when both are present", func() {
			// given
			conf := &api.Conf{
				LocalityAwareness: &api.LocalityAwareness{
					LocalZone: &api.LocalZone{
						AffinityTags: pointer.To([]api.AffinityTag{
							{Key: "region"},
						}),
					},
				},
			}

			// given: inbound tag is present alongside a conflicting pod label
			inboundTags := mesh_proto.MultiValueTagSet{
				"region": {"eu-west": true},
			}
			podLabels := map[string]string{
				"region": "us-east",
			}

			// when
			localGroups, _ := v1alpha1.GetLocalityGroups(conf, inboundTags, podLabels, "local-zone")

			// then: inbound tag value wins
			Expect(localGroups).To(HaveLen(1))
			Expect(localGroups[0].Value).To(Equal("eu-west"))
		})
	})
})
