package universal_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	store_config "github.com/kumahq/kuma/pkg/config/core/resources/store"
	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/tokens/builtin"
	builtin_issuer "github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	"github.com/kumahq/kuma/pkg/xds/auth"
	"github.com/kumahq/kuma/pkg/xds/auth/universal"
)

var _ = Describe("Authentication flow", func() {
	var dpTokenIssuer builtin_issuer.DataplaneTokenIssuer
	var authenticator auth.Authenticator
	var resStore core_store.ResourceStore
	var resManager manager.ResourceManager
	var ctx context.Context

	dpRes := *samples.DataplaneWebBuilder().
		AddInboundOfService("web-api").
		Build()

	BeforeEach(func() {
		ctx = context.Background()
		resStore = memory.NewStore()
		resManager = manager.NewResourceManager(resStore)

		Expect(resManager.Create(ctx, core_mesh.NewMeshResource(), core_store.CreateByKey("default", model.NoMesh))).To(Succeed())
		Expect(resManager.Create(ctx, core_mesh.NewMeshResource(), core_store.CreateByKey("demo", model.NoMesh))).To(Succeed())
		Expect(resManager.Create(ctx, core_mesh.NewMeshResource(), core_store.CreateByKey("demo-2", model.NoMesh))).To(Succeed())

		dpTokenIssuer = builtin.NewDataplaneTokenIssuer(resManager)
		dataplaneValidator, err := builtin.NewDataplaneTokenValidator(resManager, store_config.MemoryStore, dp_server.DpTokenValidatorConfig{
			UseSecrets: true,
		})
		Expect(err).ToNot(HaveOccurred())
		zoneTokenValidator, err := builtin.NewZoneTokenValidator(resManager, false, store_config.MemoryStore, dp_server.ZoneTokenValidatorConfig{
			UseSecrets: true,
		})
		Expect(err).ToNot(HaveOccurred())
		authenticator = universal.NewAuthenticator(dataplaneValidator, zoneTokenValidator, "zone-1")

		signingKeyManager := tokens.NewMeshedSigningKeyManager(resManager, system.DataplaneTokenSigningKey("default"), "default")
		Expect(signingKeyManager.CreateDefaultSigningKey(ctx)).To(Succeed())
		signingKeyManager = tokens.NewMeshedSigningKeyManager(resManager, system.DataplaneTokenSigningKey("demo"), "demo")
		Expect(signingKeyManager.CreateDefaultSigningKey(ctx)).To(Succeed())

		err = resStore.Create(ctx, &dpRes, core_store.CreateByKey("dp-1", "default"))
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
			credential, err := dpTokenIssuer.Generate(ctx, given.id, 24*time.Hour)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = authenticator.Authenticate(ctx, given.dpRes, credential)

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
			token, err := dpTokenIssuer.Generate(ctx, given.id, 24*time.Hour)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = authenticator.Authenticate(ctx, given.dpRes, token)

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
			err:   "could not parse token. kuma-cp runs with an in-memory database and its state isn't preserved between restarts. Keep in mind that an in-memory database cannot be used with multiple instances of the control plane: crypto/rsa: verification error",
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
		err := authenticator.Authenticate(ctx, &dpRes, "this-is-not-valid-jwt-token")

		// then
		Expect(err.Error()).To(ContainSubstring("could not parse token. kuma-cp runs with an in-memory database and its state isn't preserved between restarts." +
			" Keep in mind that an in-memory database cannot be used with multiple instances of the control plane: token contains an invalid number of segments"))
	})

	It("should throw an error when signing key used for validation is different than for generation", func() {
		// given
		credential, err := dpTokenIssuer.Generate(ctx, builtin_issuer.DataplaneIdentity{
			Name: "dp-1",
			Mesh: "default",
		}, 24*time.Hour)
		Expect(err).ToNot(HaveOccurred())

		// and new signing key
		signingKeyManager := tokens.NewMeshedSigningKeyManager(resManager, system.DataplaneTokenSigningKey("default"), "default")
		Expect(resManager.DeleteAll(ctx, &system.SecretResourceList{})).To(Succeed())
		Expect(signingKeyManager.CreateDefaultSigningKey(ctx)).To(Succeed())

		// when
		err = authenticator.Authenticate(ctx, &dpRes, credential)

		// then
		Expect(err.Error()).To(ContainSubstring("could not parse token. kuma-cp runs with an in-memory database and its state isn't preserved between restarts." +
			" Keep in mind that an in-memory database cannot be used with multiple instances of the control plane: crypto/rsa: verification error"))
	})

	It("should throw an error when signing key is not found", func() {
		// when
		_, err := dpTokenIssuer.Generate(ctx, builtin_issuer.DataplaneIdentity{
			Mesh: "demo-2",
		}, 24*time.Hour)

		// then
		Expect(err.Error()).To(ContainSubstring(`there is no signing key`))
	})

	Context("Zone Ingress", func() {
		ziRes := core_mesh.ZoneIngressResource{
			Meta: &test_model.ResourceMeta{
				Mesh: "zi-1",
			},
			Spec: &mesh_proto.ZoneIngress{
				Networking: &mesh_proto.ZoneIngress_Networking{
					Address: "127.0.0.1",
				},
			},
		}

		var zoneTokenIssuer zone.TokenIssuer

		BeforeEach(func() {
			err := resStore.Create(ctx, &ziRes, core_store.CreateByKey("zi-1", model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			zoneKeyManager := tokens.NewSigningKeyManager(resManager, system.ZoneTokenSigningKeyPrefix)
			Expect(zoneKeyManager.CreateDefaultSigningKey(ctx)).To(Succeed())
			zoneTokenIssuer = builtin.NewZoneTokenIssuer(resManager)
		})

		It("should authenticate zone ingress with zone token", func() {
			// given
			identity := zone.Identity{
				Zone:  "zone-1",
				Scope: []string{zone.IngressScope},
			}
			token, err := zoneTokenIssuer.Generate(ctx, identity, time.Hour)
			Expect(err).ToNot(HaveOccurred())

			// when
			err = authenticator.Authenticate(ctx, &ziRes, token)

			// then
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
