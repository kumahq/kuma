package topology

import (
	"net"

	"github.com/kumahq/kuma/pkg/core/dns"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var _ = Describe("Resolve Dataplane address", func() {
	dns.LookupIP = func(s string) ([]net.IP, error) {
		if s == "example.com" {
			return []net.IP{net.ParseIP("192.168.0.1")}, nil
		}
		return nil, errors.New("can't resolve host name")
	}

	It("should resolve if networking.address is domain name", func() {
		dp := &mesh.DataplaneResource{Spec: mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{Address: "example.com"}},
		}
		err := ResolveAddress(dp)
		Expect(err).ToNot(HaveOccurred())
		Expect(dp.Spec.Networking.Address).To(Equal("192.168.0.1"))
	})
})
