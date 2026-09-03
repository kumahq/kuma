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

	It("should reject a resource that is not a Dataplane", func() {
		// given
		mesh := core_mesh.MeshResource{Spec: &mesh_proto.Mesh{}}

		// when
		err := authenticator.Authenticate(context.Background(), &mesh, "some-random-token")

		// then
		Expect(err).To(MatchError("no matching authenticator for Mesh resource"))
	})
})
