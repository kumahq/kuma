package universal_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	store_config "github.com/kumahq/kuma/v3/pkg/config/core/resources/store"
	dp_server "github.com/kumahq/kuma/v3/pkg/config/dp-server"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/tokens"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	"github.com/kumahq/kuma/v3/pkg/tokens/builtin"
	builtin_issuer "github.com/kumahq/kuma/v3/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/v3/pkg/xds/auth"
	"github.com/kumahq/kuma/v3/pkg/xds/auth/universal"
)

var _ = Describe("Authentication flow", func() {
	var dpTokenIssuer builtin_issuer.DataplaneTokenIssuer
	var authenticator auth.Authenticator
	var resStore core_store.ResourceStore
	var resManager manager.ResourceManager
	var ctx context.Context

	dpRes := *samples.DataplaneWebBuilder().
		Build()

	var dpResWithWorkload core_mesh.DataplaneResource

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
		authenticator = universal.NewAuthenticator(dataplaneValidator, resManager, config_core.UniversalEnvironment)

		signingKeyManager := tokens.NewMeshedSigningKeyManager(resManager, system.DataplaneTokenSigningKey("default"), "default")
		Expect(signingKeyManager.CreateDefaultSigningKey(ctx)).To(Succeed())
		signingKeyManager = tokens.NewMeshedSigningKeyManager(resManager, system.DataplaneTokenSigningKey("demo"), "demo")
		Expect(signingKeyManager.CreateDefaultSigningKey(ctx)).To(Succeed())

		err = resStore.Create(ctx, &dpRes, core_store.CreateByKey("dp-1", "default"), core_store.CreateWithLabels(map[string]string{
			"team":                 "web",
			"env":                  "prod",
			"kuma.io/display-name": "web-api",
		}))
		Expect(err).ToNot(HaveOccurred())

		dpResWithWorkload = *samples.DataplaneWebBuilder().
			WithName("dp-with-workload").
			WithLabels(map[string]string{metadata.KumaWorkload: "backend", "team": "web"}).
			Build()
		err = resStore.Create(ctx, &dpResWithWorkload, core_store.CreateByKey("dp-with-workload", "default"), core_store.CreateWithLabels(map[string]string{metadata.KumaWorkload: "backend", "team": "web"}))
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
					"team": {
						"web": true,
					},
				},
			},
			dpRes: &dpRes,
		}),
		Entry("should auth with token bound to workload", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh:     "default",
				Workload: "backend",
			},
			dpRes: &dpResWithWorkload,
		}),
		Entry("should auth with a token bound to one of several allowed tag values", testCase{
			// The token allows either "web" or "web-api"; dpRes's actual
			// kuma.io/display-name label is "web-api", so it still matches.
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Tags: map[string]map[string]bool{
					"kuma.io/display-name": {
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
			err:   "could not parse token. kuma-cp runs with an in-memory database and its state isn't preserved between restarts. Keep in mind that an in-memory database cannot be used with multiple instances of the control plane: token signature is invalid: crypto/rsa: verification error",
		}),
		Entry("on token with different tags", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Tags: map[string]map[string]bool{
					"team": {
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
		Entry("on token bound to multiple tags where one does not match", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Tags: map[string]map[string]bool{
					"team": {
						"web": true,
					},
					"env": {
						"staging": true,
					},
				},
			},
			dpRes: &dpRes,
			err:   `which is not allowed with this token. Allowed values in token are ["staging"]`,
		}),
		Entry("on token with different workload", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh:     "default",
				Workload: "frontend",
			},
			dpRes: &dpResWithWorkload,
			err:   `dataplane workload "backend" is not allowed with this token. Allowed workload in token is "frontend"`,
		}),
		Entry("on token with workload when dataplane has no workload label", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh:     "default",
				Workload: "backend",
			},
			dpRes: &dpRes,
			err:   "dataplane has no workload label required by the token",
		}),
	)

	// A MeshIdentity whose SPIFFE ID path template references the workload makes
	// the kuma.io/workload label the proxy's identity, so a token that is not
	// bound to that workload must not be accepted when the dataplane declares a
	// workload label. MeshIdentities whose path does not reference the workload
	// (the Kubernetes ns/sa template, custom non-workload templates) leave
	// unbound tokens valid, since the workload label is not the identity.
	Context("workload label as the proxy's SPIFFE identity", func() {
		// tagsOnlyToken returns a token bound to tags but not to a workload; it
		// matches both dpRes and dpResWithWorkload.
		tagsOnlyToken := func() string {
			token, err := dpTokenIssuer.Generate(ctx, builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Tags: map[string]map[string]bool{
					"team": {
						"web": true,
					},
				},
			}, 24*time.Hour)
			Expect(err).ToNot(HaveOccurred())
			return token
		}

		Context("when the matched MeshIdentity derives the identity from the workload label", func() {
			BeforeEach(func() {
				// Empty selector matches every dataplane and, with no custom
				// SpiffeID, uses the default universal template
				// (/workload/{{ .Workload }}).
				Expect(builders.MeshIdentity().WithName("identity-1").Create(resStore)).To(Succeed())
			})

			It("should reject a tags-only token when the dataplane declares a workload label", func() {
				// when a token that is not bound to a workload is presented for a
				// dataplane that declares a workload label
				err := authenticator.Authenticate(ctx, &dpResWithWorkload, tagsOnlyToken())

				// then it is rejected
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("requires a workload-bound token"))
			})

			It("should accept a workload-bound token matching the dataplane workload label", func() {
				// given a token bound to the workload
				token, err := dpTokenIssuer.Generate(ctx, builtin_issuer.DataplaneIdentity{
					Mesh:     "default",
					Workload: "backend",
				}, 24*time.Hour)
				Expect(err).ToNot(HaveOccurred())

				// when / then
				Expect(authenticator.Authenticate(ctx, &dpResWithWorkload, token)).To(Succeed())
			})

			It("should accept a tags-only token when the dataplane has no workload label", func() {
				// when a tags-only token is presented for a dataplane that does
				// not declare a workload label, there is no workload label to bind
				Expect(authenticator.Authenticate(ctx, &dpRes, tagsOnlyToken())).To(Succeed())
			})
		})

		DescribeTable("should accept a tags-only token when the identity does not derive from the workload label",
			func(spiffeIDPath string) {
				// given a MeshIdentity whose SPIFFE path does not reference the workload
				Expect(builders.MeshIdentity().
					WithName("identity-1").
					WithSpiffeID("{{ .Mesh }}.{{ .Zone }}.mesh.local", spiffeIDPath).
					Create(resStore)).To(Succeed())

				// when a token that is not bound to a workload is presented for a
				// dataplane that declares a workload label
				err := authenticator.Authenticate(ctx, &dpResWithWorkload, tagsOnlyToken())

				// then it is accepted, because the workload label is not the identity
				Expect(err).ToNot(HaveOccurred())
			},
			Entry("Kubernetes ns/sa template", "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"),
			Entry("custom non-workload template", "/custom/{{ .Mesh }}"),
		)

		It("should accept a tags-only token with a workload label when no MeshIdentity matches", func() {
			// given no MeshIdentity in the mesh, the identity does not derive from
			// the workload label, so a tags-only token stays valid
			Expect(authenticator.Authenticate(ctx, &dpResWithWorkload, tagsOnlyToken())).To(Succeed())
		})
	})

	It("should throw an error on invalid token", func() {
		// when
		err := authenticator.Authenticate(ctx, &dpRes, "this-is-not-valid-jwt-token")

		// then
		Expect(err.Error()).To(ContainSubstring("could not parse token. kuma-cp runs with an in-memory database and its state isn't preserved between restarts." +
			" Keep in mind that an in-memory database cannot be used with multiple instances of the control plane: token is malformed: token contains an invalid number of segments"))
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
			" Keep in mind that an in-memory database cannot be used with multiple instances of the control plane: token signature is invalid: crypto/rsa: verification error"))
	})

	It("should throw an error when signing key is not found", func() {
		// when
		_, err := dpTokenIssuer.Generate(ctx, builtin_issuer.DataplaneIdentity{
			Mesh: "demo-2",
		}, 24*time.Hour)

		// then
		Expect(err.Error()).To(ContainSubstring(`there is no signing key`))
	})

	Context("zone proxy", func() {
		zoneProxyRes := core_mesh.DataplaneResource{
			Meta: &test_model.ResourceMeta{
				Name: "zone-proxy-1",
				Mesh: "default",
			},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "127.0.0.1",
					Listeners: []*mesh_proto.Dataplane_Networking_Listener{{
						Type:    mesh_proto.Dataplane_Networking_Listener_ZoneIngress,
						Address: "127.0.0.1",
						Port:    10001,
					}},
				},
			},
		}

		BeforeEach(func() {
			Expect(resStore.Create(ctx, &zoneProxyRes, core_store.CreateByKey("zone-proxy-1", "default"))).To(Succeed())
		})

		It("should authenticate a zone proxy Dataplane with a dataplane token", func() {
			// given
			token, err := dpTokenIssuer.Generate(ctx, builtin_issuer.DataplaneIdentity{
				Name: "zone-proxy-1",
				Mesh: "default",
			}, 24*time.Hour)
			Expect(err).ToNot(HaveOccurred())

			// when / then
			Expect(authenticator.Authenticate(ctx, &zoneProxyRes, token)).To(Succeed())
		})
	})
})
