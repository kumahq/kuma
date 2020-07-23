package universal_test

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/sds/auth"
	"github.com/kumahq/kuma/pkg/sds/auth/universal"
	"github.com/kumahq/kuma/pkg/sds/server"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Noop Authenticator", func() {

	var authenticator auth.Authenticator
	var resStore store.ResourceStore

	BeforeEach(func() {
		resStore = memory.NewStore()
		authenticator = universal.NewNoopAuthenticator(server.DefaultDataplaneResolver(manager.NewResourceManager(resStore)))
	})

	It("should allow with any token for existing dataplane", func() {
		// given
		id := xds.ProxyId{
			Mesh: "example",
			Name: "dp-1",
		}

		dpRes := core_mesh.DataplaneResource{
			Spec: mesh_proto.Dataplane{
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
		err := resStore.Create(context.Background(), &dpRes, store.CreateBy(id.ToResourceKey()))
		Expect(err).ToNot(HaveOccurred())

		// when
		authIdentity, err := authenticator.Authenticate(context.Background(), id, "some-random-token")

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(authIdentity.Services[0]).To(Equal("web"))
		Expect(authIdentity.Mesh).To(Equal(id.Mesh))
	})

	It("should throw an error when dataplane is not present in CP", func() {
		// given
		id := xds.ProxyId{
			Mesh: "example",
			Name: "dp-1",
		}

		// when
		_, err := authenticator.Authenticate(context.Background(), id, "some-random-token")

		// then
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(`unable to find Dataplane for proxy "example.dp-1": Resource not found: type="Dataplane" name="dp-1" mesh="example"`))
	})
})
