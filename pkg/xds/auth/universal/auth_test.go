package universal_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/tokens/builtin"
	builtin_issuer "github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/xds/auth"
	"github.com/kumahq/kuma/pkg/xds/auth/universal"
)

var _ = Describe("Authentication flow", func() {
	var issuer builtin_issuer.DataplaneTokenIssuer
	var authenticator auth.Authenticator
	var resStore core_store.ResourceStore
	var ctx context.Context

	dpRes := core_mesh.DataplaneResource{
		Meta: &test_model.ResourceMeta{
			Mesh: "dp-1",
			Name: "default",
		},
		Spec: &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "127.0.0.1",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port:        8080,
						ServicePort: 8081,
						Tags: map[string]string{
							"kuma.io/service":  "web",
							"kuma.io/protocol": "http",
						},
					},
					{
						Port:        8090,
						ServicePort: 8091,
						Tags: map[string]string{
							"kuma.io/service":  "web-api",
							"kuma.io/protocol": "http",
						},
					},
				},
			},
		},
	}

	BeforeEach(func() {
		ctx = context.Background()
		resStore = memory.NewStore()
		resManager := manager.NewResourceManager(resStore)

		Expect(resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey("default", model.NoMesh))).To(Succeed())
		Expect(resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey("demo", model.NoMesh))).To(Succeed())
		Expect(resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey("demo-2", model.NoMesh))).To(Succeed())

		dataplaneValidator := builtin.NewDataplaneTokenValidator(resManager)
		zoneIngressValidator := builtin.NewZoneIngressTokenValidator(resManager)
		zoneTokenValidator := builtin.NewZoneTokenValidator(resManager, config_core.Global)
		issuer = builtin.NewDataplaneTokenIssuer(resManager)
		authenticator = universal.NewAuthenticator(dataplaneValidator, zoneIngressValidator, zoneTokenValidator, "zone-1")

		signingKeyManager := tokens.NewMeshedSigningKeyManager(resManager, builtin_issuer.DataplaneTokenSigningKeyPrefix("default"), "default")
		Expect(signingKeyManager.CreateDefaultSigningKey(ctx)).To(Succeed())
		signingKeyManager = tokens.NewMeshedSigningKeyManager(resManager, builtin_issuer.DataplaneTokenSigningKeyPrefix("demo"), "demo")
		Expect(signingKeyManager.CreateDefaultSigningKey(ctx)).To(Succeed())

		err := resStore.Create(context.Background(), &dpRes, core_store.CreateByKey("dp-1", "default"))
		Expect(err).ToNot(HaveOccurred())
	})

	type testCase struct {
		id    builtin_issuer.DataplaneIdentity
		dpRes *core_mesh.DataplaneResource
		err   string
	}
	DescribeTable("should correctly authenticate dataplane",
		func(given testCase) {
			// when
			credential, err := issuer.Generate(ctx, given.id, 24*time.Hour)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = authenticator.Authenticate(context.Background(), given.dpRes, credential)

			// then
			Expect(err).ToNot(HaveOccurred())
		},
		Entry("should auth with token bound to mesh", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "default",
			},
			dpRes: &dpRes,
		}),
		Entry("should auth with token bound to mesh and name", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Name: "dp-1",
				Mesh: "default",
			},
			dpRes: &dpRes,
		}),
		Entry("should auth with token bound to mesh and tags", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Tags: map[string]map[string]bool{
					"kuma.io/service": {
						"web":     true,
						"web-api": true,
					},
				},
			},
			dpRes: &dpRes,
		}),
	)

	DescribeTable("should fail auth",
		func(given testCase) {
			// when
			token, err := issuer.Generate(ctx, given.id, 24*time.Hour)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = authenticator.Authenticate(context.Background(), given.dpRes, token)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(given.err))
		},
		Entry("on token with different name", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Name: "dp-2",
			},
			dpRes: &dpRes,
			err:   "proxy name from requestor: dp-1 is different than in token: dp-2",
		}),
		Entry("on token with different mesh", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "demo",
				Name: "dp-1",
			},
			dpRes: &dpRes,
			err:   "could not parse token: crypto/rsa: verification error",
		}),
		Entry("on token with different tags", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Tags: map[string]map[string]bool{
					"kuma.io/service": {
						"backend": true,
					},
				},
			},
			dpRes: &dpRes,
			err:   `which is not allowed with this token. Allowed values in token are ["backend"]`,
		}),
		Entry("on token with tag that is absent in dataplane", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Tags: map[string]map[string]bool{
					"kuma.io/zone": {
						"east": true,
					},
				},
			},
			dpRes: &dpRes,
			err:   `dataplane has no tag "kuma.io/zone" required by the token`,
		}),
		Entry("on token with missing one tag value", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Tags: map[string]map[string]bool{
					"kuma.io/service": {
						"web": true,
						// "web-api": true valid token should have also web-api
					},
				},
			},
			dpRes: &dpRes,
			err:   `which is not allowed with this token. Allowed values in token are ["web"]`, // web and web-api order is not stable
		}),
	)

	It("should throw an error on invalid token", func() {
		// when
		err := authenticator.Authenticate(context.Background(), &dpRes, "this-is-not-valid-jwt-token")

		// then
		Expect(err).To(MatchError("could not parse token: token contains an invalid number of segments"))
	})

	It("should throw an error when signing key is not found", func() {
		// when
		_, err := issuer.Generate(ctx, builtin_issuer.DataplaneIdentity{
			Mesh: "demo-2",
		}, 24*time.Hour)

		// then
		Expect(err.Error()).To(ContainSubstring(`there is no signing key`))
	})
})
