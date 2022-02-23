package universal_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/xds/auth"
	"github.com/kumahq/kuma/pkg/xds/auth/universal"
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
							Tags: map[string]string{
								"kuma.io/service": "web",
							},
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

	It("should allow with any token for any zone ingress", func() {
		// given
		zoneIngress := core_mesh.ZoneIngressResource{
			Spec: &mesh_proto.ZoneIngress{
				Zone: "zone-1",
				Networking: &mesh_proto.ZoneIngress_Networking{
					Address:           "127.0.0.1",
					AdvertisedAddress: "192.168.0.1",
					Port:              10001,
					AdvertisedPort:    10001,
				},
				AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
					{
						Tags: map[string]string{
							"kuma.io/service": "web",
						},
						Instances: 1,
						Mesh:      "default",
					},
				},
			},
		}

		// when
		err := authenticator.Authenticate(context.Background(), &zoneIngress, "some-random-token")

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should allow with any token for any zone egress", func() {
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
		Expect(err).ToNot(HaveOccurred())
	})
})
