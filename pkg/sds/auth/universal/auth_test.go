package universal_test

import (
	"context"
	"github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/sds/auth"
	"github.com/Kong/kuma/pkg/sds/auth/universal"
	"github.com/Kong/kuma/pkg/sds/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Authentication flow", func() {
	var privateKey = []byte("testPrivateKey")

	generator := universal.NewCredentialGenerator(privateKey)
	var authenticator auth.Authenticator
	var resStore store.ResourceStore

	BeforeEach(func() {
		resStore = memory.NewStore()
		authenticator = universal.NewAuthenticator(privateKey, server.DefaultDataplaneResolver(manager.NewResourceManager(manager.NewResourceManager(resStore))))
	})

	It("should correctly authenticate dataplane", func() {
		// given
		id := xds.ProxyId{
			Mesh:      "example",
			Namespace: "default",
			Name:      "dp-1",
		}

		dpRes := core_mesh.DataplaneResource{
			Spec: v1alpha1.Dataplane{
				Networking: &v1alpha1.Dataplane_Networking{
					Inbound: []*v1alpha1.Dataplane_Networking_Inbound{
						{
							Interface: "127.0.0.1:8080:8081",
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
		credential, err := generator.Generate(id)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		authIdentity, err := authenticator.Authenticate(context.Background(), id, credential)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(authIdentity.Service).To(Equal("web"))
		Expect(authIdentity.Mesh).To(Equal(id.Mesh))
	})

	It("should throw an error on invalid token", func() {
		// when
		id := xds.ProxyId{
			Mesh:      "default",
			Namespace: "default",
			Name:      "dp1",
		}
		_, err := authenticator.Authenticate(context.Background(), id, "this-is-not-valid-jwt-token")

		// then
		Expect(err).To(MatchError("could not parse token: token contains an invalid number of segments"))
	})

	It("should throw an error on token with different name", func() {
		// when
		generateId := xds.ProxyId{
			Mesh:      "default",
			Namespace: "default",
			Name:      "different-name-than-dp1",
		}
		token, err := generator.Generate(generateId)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		authId := xds.ProxyId{
			Mesh:      "default",
			Namespace: "default",
			Name:      "dp1",
		}
		_, err = authenticator.Authenticate(context.Background(), authId, token)

		// then
		Expect(err).To(MatchError("proxy name from requestor is different than in token. Expected dp1 got different-name-than-dp1"))
	})

	It("should throw an error on token with different mesh", func() {
		// when
		generateId := xds.ProxyId{
			Mesh:      "different-mesh-than-default",
			Namespace: "default",
			Name:      "dp1",
		}
		token, err := generator.Generate(generateId)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		authId := xds.ProxyId{
			Mesh:      "default",
			Namespace: "default",
			Name:      "dp1",
		}
		_, err = authenticator.Authenticate(context.Background(), authId, token)

		// then
		Expect(err).To(MatchError("proxy mesh from requestor is different than in token. Expected default got different-mesh-than-default"))
	})

	It("should throw an error when dataplane is not present in CP", func() {
		// given
		id := xds.ProxyId{
			Mesh:      "default",
			Namespace: "default",
			Name:      "non-existent-dp",
		}

		// when
		token, err := generator.Generate(id)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = authenticator.Authenticate(context.Background(), id, token)

		// then
		Expect(err).To(MatchError(`unable to find Dataplane for proxy {"default" "default" "non-existent-dp"}: Resource not found: type="Dataplane" namespace="default" name="non-existent-dp" mesh="default"`))
	})
})
