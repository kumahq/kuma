package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshzoneaddress_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshzoneaddress/api/v1alpha1"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
)

func meshZoneAddress(version string, address string, port int32) *meshzoneaddress_api.MeshZoneAddressResource {
	return &meshzoneaddress_api.MeshZoneAddressResource{
		Meta: &test_model.ResourceMeta{
			Mesh:    "default",
			Name:    "zone-proxy-east",
			Version: version,
		},
		Spec: &meshzoneaddress_api.MeshZoneAddress{
			Address: address,
			Port:    port,
		},
	}
}

var _ = Describe("Hash", func() {
	It("should be stable for identical resources", func() {
		Expect(meshZoneAddress("1", "10.0.0.1", 10001).Hash()).
			To(Equal(meshZoneAddress("1", "10.0.0.1", 10001).Hash()))
	})

	// the mesh context stores the address after DNS resolution, so a load balancer
	// hostname keeps the same resourceVersion while resolving to a new IP
	It("should change when the address changes", func() {
		Expect(meshZoneAddress("1", "10.0.0.1", 10001).Hash()).
			ToNot(Equal(meshZoneAddress("1", "10.0.0.2", 10001).Hash()))
	})

	It("should change when the port changes", func() {
		Expect(meshZoneAddress("1", "10.0.0.1", 10001).Hash()).
			ToNot(Equal(meshZoneAddress("1", "10.0.0.1", 10002).Hash()))
	})

	It("should change when the version changes", func() {
		Expect(meshZoneAddress("1", "10.0.0.1", 10001).Hash()).
			ToNot(Equal(meshZoneAddress("2", "10.0.0.1", 10001).Hash()))
	})

	It("should not collide when the address and port boundary shifts", func() {
		Expect(meshZoneAddress("1", "10.0.0.1", 23).Hash()).
			ToNot(Equal(meshZoneAddress("1", "10.0.0.12", 3).Hash()))
	})

	It("should not collide for an IPv6 address ending in digits", func() {
		Expect(meshZoneAddress("1", "fd00::1", 23).Hash()).
			ToNot(Equal(meshZoneAddress("1", "fd00::12", 3).Hash()))
	})

	It("should be computed for a nil spec", func() {
		withNilSpec := &meshzoneaddress_api.MeshZoneAddressResource{
			Meta: &test_model.ResourceMeta{Mesh: "default", Name: "zone-proxy-east"},
		}

		Expect(withNilSpec.Hash()).ToNot(BeEmpty())
	})
})
