package mesh_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

func zoneIngress(address string, advertisedAddress string) *core_mesh.ZoneIngressResource {
	return &core_mesh.ZoneIngressResource{
		Meta: &test_model.ResourceMeta{
			Mesh:    "default",
			Name:    "zone-ingress-east",
			Version: "1",
		},
		Spec: &mesh_proto.ZoneIngress{
			Networking: &mesh_proto.ZoneIngress_Networking{
				Address:           address,
				AdvertisedAddress: advertisedAddress,
			},
		},
	}
}

var _ = Describe("ZoneIngress Hash", func() {
	It("should be stable for identical resources", func() {
		Expect(zoneIngress("10.0.0.1", "192.168.0.1").Hash()).
			To(Equal(zoneIngress("10.0.0.1", "192.168.0.1").Hash()))
	})

	It("should change when the advertised address changes", func() {
		Expect(zoneIngress("10.0.0.1", "192.168.0.1").Hash()).
			ToNot(Equal(zoneIngress("10.0.0.1", "192.168.0.2").Hash()))
	})

	// the address and the advertised address are written back to back, so without
	// length prefixes the pair boundary can shift without changing the hash
	It("should not collide when the address boundary shifts", func() {
		Expect(zoneIngress("10.0.0.1", "1.2.3.4").Hash()).
			ToNot(Equal(zoneIngress("10.0.0.11", ".2.3.4").Hash()))
	})
})
