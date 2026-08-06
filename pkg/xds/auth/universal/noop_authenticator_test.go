package universal_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/xds/auth"
	"github.com/kumahq/kuma/v3/pkg/xds/auth/universal"
)

var _ = Describe("Noop Authenticator", func() {
	var authenticator auth.Authenticator

	BeforeEach(func() {
		authenticator = universal.NewNoopAuthenticator()
	})

	It("should allow with any token for any dataplane", func() {
		// given
		dpRes := core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "127.0.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        8080,
							ServicePort: 8081,
						},
					},
				},
			},
		}

		// when
		err := authenticator.Authenticate(context.Background(), &dpRes, "some-random-token")

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should reject a zone ingress", func() {
		// given
		zoneIngress := core_mesh.ZoneIngressResource{
			Spec: &mesh_proto.ZoneIngress{
				Zone: "zone-1",
				Networking: &mesh_proto.ZoneIngress_Networking{
					Address: "127.0.0.1",
					Port:    10001,
				},
			},
		}

		// when
		err := authenticator.Authenticate(context.Background(), &zoneIngress, "some-random-token")

		// then
		Expect(err).To(MatchError("no matching authenticator for ZoneIngress resource"))
	})

	It("should reject a zone egress", func() {
		// given
		zoneEgress := core_mesh.ZoneEgressResource{
			Spec: &mesh_proto.ZoneEgress{
				Zone: "zone-1",
				Networking: &mesh_proto.ZoneEgress_Networking{
					Address: "127.0.0.1",
					Port:    10002,
				},
			},
		}

		// when
		err := authenticator.Authenticate(context.Background(), &zoneEgress, "some-random-token")

		// then
		Expect(err).To(MatchError("no matching authenticator for ZoneEgress resource"))
	})
})
