package topology

import (
	"net"

	"github.com/kumahq/kuma/pkg/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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
		return nil, errors.New("can't resolve host name")
	}

	Context("ResolveAddress", func() {
		It("should resolve if networking.address is domain name", func() {
			// given
			dp := &mesh.DataplaneResource{Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{Address: "example.com"}},
			}

			// when
			resolvedDp, err := ResolveAddress(lif, dp)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(resolvedDp.Spec.Networking.Address).To(Equal("192.168.0.1"))
			// and original DP is not modified
			Expect(dp.Spec.Networking.Address).To(Equal("example.com"))
		})
	})

	Context("ResolveAddresses", func() {
		It("should resolve addresses for all dataplanes", func() {
			given := []*mesh.DataplaneResource{
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "example.com"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "example-0.com"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "example-1.com"}}},
			}
			expected := []*mesh.DataplaneResource{
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "192.168.0.1"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "192.168.1.0"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "192.168.1.1"}}},
			}

			actual := ResolveAddresses(core.Log, lif, given)
			Expect(actual).To(HaveLen(3))
			Expect(actual).To(Equal(expected))
		})
		It("should skip dataplane if unable to resolve domain name", func() {
			given := []*mesh.DataplaneResource{
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "example.com"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "unresolvable.com"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "example-1.com"}}},
			}
			expected := []*mesh.DataplaneResource{
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "192.168.0.1"}}},
				{Spec: &mesh_proto.Dataplane{Networking: &mesh_proto.Dataplane_Networking{Address: "192.168.1.1"}}},
			}
			actual := ResolveAddresses(core.Log, lif, given)
			Expect(actual).To(HaveLen(2))
			Expect(actual).To(Equal(expected))
		})
	})
})
