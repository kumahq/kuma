package topology_test

import (
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/xds/topology"
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
		return nil, errors.New("can't resolve host name")
	}

	Context("ResolveAddress", func() {
		It("should resolve if networking.address is domain name", func() {
			// given
			dp := &mesh.DataplaneResource{Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{Address: "example.com", AdvertisedAddress: "advertise.example.com"}},
			}

			// when
			resolvedDp, err := topology.ResolveAddress(lif, dp)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resolvedDp.Spec.Networking.Address).To(Equal("192.168.0.1"))
			Expect(resolvedDp.Spec.Networking.AdvertisedAddress).To(Equal("192.0.2.1"))
			// and original DP is not modified
			Expect(dp.Spec.Networking.Address).To(Equal("example.com"))
			Expect(dp.Spec.Networking.AdvertisedAddress).To(Equal("advertise.example.com"))
		})
	})

	Context("ResolveAddresses", func() {
		It("should resolve addresses for all dataplanes", func() {
			given := []*mesh.DataplaneResource{
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "example.com"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "example-0.com"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "example-1.com"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "example-2.com", AdvertisedAddress: "advertise-2.example.com"}}},
			}
			expected := []*mesh.DataplaneResource{
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "192.168.0.1"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "192.168.1.0"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "192.168.1.1"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "192.168.1.2", AdvertisedAddress: "192.0.2.2"}}},
			}

			actual := topology.ResolveAddresses(core.Log, lif, given)
			Expect(actual).To(HaveLen(4))
			Expect(actual).To(Equal(expected))
		})
		It("should skip dataplane if unable to resolve domain name", func() {
			given := []*mesh.DataplaneResource{
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "example.com"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "unresolvable.com"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "example-1.com"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "example-2.com", AdvertisedAddress: "advertise-2.example.com"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "example-3.com", AdvertisedAddress: "abc.example.com"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{AdvertisedAddress: "advertise-2.example.com"}}},
			}
			expected := []*mesh.DataplaneResource{
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "192.168.0.1"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "192.168.1.1"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "192.168.1.2", AdvertisedAddress: "192.0.2.2"}}},
			}
			actual := topology.ResolveAddresses(core.Log, lif, given)
			Expect(actual).To(HaveLen(3))
			Expect(actual).To(Equal(expected))
		})
	})
})
