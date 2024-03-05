package topology_test

import (
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
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
})
