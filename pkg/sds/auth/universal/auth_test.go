package universal_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/sds/auth"
	"github.com/Kong/kuma/pkg/sds/auth/universal"
	"github.com/Kong/kuma/pkg/sds/server"
	builtin_issuer "github.com/Kong/kuma/pkg/tokens/builtin/issuer"
)

var _ = Describe("Authentication flow", func() {
	var privateKey = []byte("testPrivateKey")

	issuer := builtin_issuer.NewDataplaneTokenIssuer(privateKey)
	var authenticator auth.Authenticator
	var resStore store.ResourceStore

	BeforeEach(func() {
		resStore = memory.NewStore()
		authenticator = universal.NewAuthenticator(
			issuer,
			server.DefaultDataplaneResolver(manager.NewResourceManager(resStore)),
		)
	})

	It("should correctly authenticate dataplane", func() {
		// given
		id := xds.ProxyId{
			Mesh: "example",
			Name: "dp-1",
		}

		dpRes := core_mesh.DataplaneResource{
			Spec: v1alpha1.Dataplane{
				Networking: &v1alpha1.Dataplane_Networking{
					Address: "127.0.0.1",
					Inbound: []*v1alpha1.Dataplane_Networking_Inbound{
						{
							Port:        8080,
							ServicePort: 8081,
							Tags: map[string]string{
								"service": "web",
							},
						},
					},
				},
			},
		}
		err := resStore.Create(context.Background(), &dpRes, store.CreateBy(id.ToResourceKey()))
		Expect(err).ToNot(HaveOccurred())

		// when
		credential, err := issuer.Generate(id)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		authIdentity, err := authenticator.Authenticate(context.Background(), id, credential)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(authIdentity.Services[0]).To(Equal("web"))
		Expect(authIdentity.Mesh).To(Equal(id.Mesh))
	})

	It("should throw an error on invalid token", func() {
		// when
		id := xds.ProxyId{
			Mesh: "default",
			Name: "dp1",
		}
		_, err := authenticator.Authenticate(context.Background(), id, "this-is-not-valid-jwt-token")

		// then
		Expect(err).To(MatchError("could not parse token: token contains an invalid number of segments"))
	})

	It("should throw an error on token with different name", func() {
		// when
		generateId := xds.ProxyId{
			Mesh: "default",
			Name: "different-name-than-dp1",
		}
		token, err := issuer.Generate(generateId)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		authId := xds.ProxyId{
			Mesh: "default",
			Name: "dp1",
		}
		_, err = authenticator.Authenticate(context.Background(), authId, token)

		// then
		Expect(err).To(MatchError("proxy name from requestor: dp1 is different than in token: different-name-than-dp1"))
	})

	It("should throw an error on token with different mesh", func() {
		// when
		generateId := xds.ProxyId{
			Mesh: "different-mesh-than-default",
			Name: "dp1",
		}
		token, err := issuer.Generate(generateId)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		authId := xds.ProxyId{
			Mesh: "default",
			Name: "dp1",
		}
		_, err = authenticator.Authenticate(context.Background(), authId, token)

		// then
		Expect(err).To(MatchError("proxy mesh from requestor: default is different than in token: different-mesh-than-default"))
	})

	It("should throw an error when dataplane is not present in CP", func() {
		// given
		id := xds.ProxyId{
			Mesh: "default",
			Name: "non-existent-dp",
		}

		// when
		token, err := issuer.Generate(id)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = authenticator.Authenticate(context.Background(), id, token)

		// then
		Expect(err).To(MatchError(`unable to find Dataplane for proxy "default.non-existent-dp": Resource not found: type="Dataplane" name="non-existent-dp" mesh="default"`))
	})
})
