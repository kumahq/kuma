package topology_test

import (
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	meshzoneaddress_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshzoneaddress/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/xds/topology"
)

var _ = Describe("Resolve Dataplane address", func() {
	lif := func(s string) ([]net.IP, error) {
		if s == "example.com" {
			return []net.IP{net.ParseIP("192.168.0.1")}, nil
		}
		if s == "example-0.com" {
			return []net.IP{net.ParseIP("192.168.1.0")}, nil
		}
		if s == "example-1.com" {
			return []net.IP{net.ParseIP("192.168.1.1")}, nil
		}
		if s == "example-2.com" {
			return []net.IP{net.ParseIP("192.168.1.2")}, nil
		}
		if s == "advertise.example.com" {
			return []net.IP{net.ParseIP("192.0.2.1")}, nil
		}
		if s == "advertise-2.example.com" {
			return []net.IP{net.ParseIP("192.0.2.2")}, nil
		}
		return nil, errors.Errorf("can't resolve hostname: %s", s)
	}

	Context("ResolveDataplaneAddress", func() {
		It("should resolve if networking.address is domain name", func() {
			// given
			dp := &mesh.DataplaneResource{
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{Address: "example.com", AdvertisedAddress: "advertise.example.com"},
				},
			}

			// when
			resolvedDp, err := topology.ResolveDataplaneAddress(lif, dp)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resolvedDp.Spec.Networking.Address).To(Equal("192.168.0.1"))
			Expect(resolvedDp.Spec.Networking.AdvertisedAddress).To(Equal("192.0.2.1"))
			// and original DP is not modified
			Expect(dp.Spec.Networking.Address).To(Equal("example.com"))
			Expect(dp.Spec.Networking.AdvertisedAddress).To(Equal("advertise.example.com"))
		})
	})

	Context("ResolveMeshZoneAddressPublicAddress", func() {
		It("should resolve if address is a domain name", func() {
			// given
			mza := &meshzoneaddress_api.MeshZoneAddressResource{
				Spec: &meshzoneaddress_api.MeshZoneAddress{Address: "example.com", Port: 10001},
			}

			// when
			resolvedMza, err := topology.ResolveMeshZoneAddressPublicAddress(lif, mza)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resolvedMza.Spec.Address).To(Equal("192.168.0.1"))
			Expect(resolvedMza.Spec.Port).To(Equal(int32(10001)))
			// and original MeshZoneAddress is not modified
			Expect(mza.Spec.Address).To(Equal("example.com"))
		})

		It("should keep the address as is when it's already an IP", func() {
			// given
			mza := &meshzoneaddress_api.MeshZoneAddressResource{
				Spec: &meshzoneaddress_api.MeshZoneAddress{Address: "192.168.0.1", Port: 10001},
			}

			// when
			resolvedMza, err := topology.ResolveMeshZoneAddressPublicAddress(lif, mza)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resolvedMza).To(BeIdenticalTo(mza))
		})

		It("should return an error if the domain name can't be resolved", func() {
			// given
			mza := &meshzoneaddress_api.MeshZoneAddressResource{
				Spec: &meshzoneaddress_api.MeshZoneAddress{Address: "unknown.com", Port: 10001},
			}

			// when
			_, err := topology.ResolveMeshZoneAddressPublicAddress(lif, mza)

			// then
			Expect(err).To(HaveOccurred())
		})
	})
})
