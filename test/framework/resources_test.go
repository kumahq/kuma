package framework

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/config/core"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
)

const legacyStandaloneMode core.CpMode = "standalone"

var _ = Describe("mesh resource cleanup", func() {
	DescribeTable("deletes resources owned by the local control plane",
		func(mode core.CpMode, origin mesh_proto.ResourceOrigin, expected bool) {
			meta := &test_model.ResourceMeta{Labels: map[string]string{}}
			if origin != "" {
				meta.Labels[mesh_proto.ResourceOriginLabel] = string(origin)
			}

			Expect(isManagedByMode(meta, mode)).To(Equal(expected))
		},
		Entry("Global owns global-origin resources", core.Global, mesh_proto.GlobalResourceOrigin, true),
		Entry("Global skips zone-synced resources", core.Global, mesh_proto.ZoneResourceOrigin, false),
		Entry("zone owns zone-origin resources", core.Zone, mesh_proto.ZoneResourceOrigin, true),
		Entry("zone skips Global-synced resources", core.Zone, mesh_proto.GlobalResourceOrigin, false),
		Entry("Global retains cleanup of unlabeled resources", core.Global, mesh_proto.ResourceOrigin(""), true),
		Entry("zone retains cleanup of unlabeled resources", core.Zone, mesh_proto.ResourceOrigin(""), true),
		Entry("standalone retains cleanup of global-origin resources", legacyStandaloneMode, mesh_proto.GlobalResourceOrigin, true),
		Entry("standalone retains cleanup of zone-origin resources", legacyStandaloneMode, mesh_proto.ZoneResourceOrigin, true),
	)
})
