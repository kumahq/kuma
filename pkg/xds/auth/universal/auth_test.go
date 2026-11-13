package universal_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	store_config "github.com/kumahq/kuma/v2/pkg/config/core/resources/store"
	dp_server "github.com/kumahq/kuma/v2/pkg/config/dp-server"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/core/tokens"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	"github.com/kumahq/kuma/v2/pkg/tokens/builtin"
	builtin_issuer "github.com/kumahq/kuma/v2/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/v2/pkg/tokens/builtin/zone"
	"github.com/kumahq/kuma/v2/pkg/xds/auth"
	"github.com/kumahq/kuma/v2/pkg/xds/auth/universal"
)

var _ = Describe("Authentication flow", func() {
	var dpTokenIssuer builtin_issuer.DataplaneTokenIssuer
	var authenticator auth.Authenticator
	var resStore core_store.ResourceStore
	var resManager manager.ResourceManager
	var ctx context.Context

	authenticate := func(res model.Resource, token string) error {
		_, err := authenticator.Authenticate(ctx, res, token)
		return err
	}

	dpRes := *samples.DataplaneWebBuilder().
		AddInboundOfService("web-api").
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
		zoneTokenValidator, err := builtin.NewZoneTokenValidator(resManager, false, store_config.MemoryStore, dp_server.ZoneTokenValidatorConfig{
			UseSecrets: true,
		})
		Expect(err).ToNot(HaveOccurred())
		authenticator = universal.NewAuthenticator(dataplaneValidator, zoneTokenValidator, resManager, config_core.UniversalEnvironment, "zone-1")

		signingKeyManager := tokens.NewMeshedSigningKeyManager(resManager, system.DataplaneTokenSigningKey("default"), "default")
		Expect(signingKeyManager.CreateDefaultSigningKey(ctx)).To(Succeed())
		signingKeyManager = tokens.NewMeshedSigningKeyManager(resManager, system.DataplaneTokenSigningKey("demo"), "demo")
		Expect(signingKeyManager.CreateDefaultSigningKey(ctx)).To(Succeed())

		err = resStore.Create(ctx, &dpRes, core_store.CreateByKey("dp-1", "default"))
		Expect(err).ToNot(HaveOccurred())

		dpResWithWorkload = *samples.DataplaneWebBuilder().
			WithName("dp-with-workload").
			WithLabels(map[string]string{metadata.KumaWorkload: "backend", "kuma.io/service": "web"}).
			AddInboundOfService("web-api").
			Build()
		err = resStore.Create(ctx, &dpResWithWorkload, core_store.CreateByKey("dp-with-workload", "default"), core_store.CreateWithLabels(map[string]string{metadata.KumaWorkload: "backend", "kuma.io/service": "web"}))
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
			_, err = authenticator.Authenticate(ctx, given.dpRes, credential)

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
		Entry("should auth with token bound to workload", testCase{
			id: builtin_issuer.DataplaneIdentity{
				Mesh:     "default",
				Workload: "backend",
			},
			dpRes: &dpResWithWorkload,
		}),
	)

	DescribeTable("should fail auth",
		func(given testCase) {
			// when
			token, err := dpTokenIssuer.Generate(ctx, given.id, 24*time.Hour)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			_, err = authenticator.Authenticate(ctx, given.dpRes, token)

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

	// On Universal the proxy declares its own labels, so every label a selected
	// MeshIdentity builds the SPIFFE ID from must be bound by the token.
	Context("labels the SPIFFE ID is built from", func() {
		generate := func(id builtin_issuer.DataplaneIdentity) string {
			token, err := dpTokenIssuer.Generate(ctx, id, 24*time.Hour)
			Expect(err).ToNot(HaveOccurred())
			return token
		}
		// tagsOnlyToken is bound to the service tag only; it matches both dpRes and
		// dpResWithWorkload.
		tagsOnlyToken := func() string {
			return generate(builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Tags: map[string]map[string]bool{"kuma.io/service": {"web": true, "web-api": true}},
			})
		}
		createIdentity := func(spiffeIDPath string) {
			Expect(builders.MeshIdentity().
				WithName("identity-1").
				WithSpiffeID("{{ .Mesh }}.{{ .Zone }}.mesh.local", spiffeIDPath).
				Create(resStore)).To(Succeed())
		}
		// A token tag is checked against the inbound tags, so a dataplane that binds
		// a label through a tag declares it on an inbound too.
		createDataplane := func(name string, labels map[string]string, inboundTags ...string) *core_mesh.DataplaneResource {
			builder := samples.DataplaneWebBuilder().WithName(name).WithLabels(labels)
			if len(inboundTags) > 0 {
				builder.AddInboundOfTags(append([]string{"kuma.io/service", "web"}, inboundTags...)...)
			}
			dp := builder.Build()
			Expect(resStore.Create(ctx, dp, core_store.CreateByKey(name, "default"), core_store.CreateWithLabels(labels))).To(Succeed())
			return dp
		}

		Context("with the default Universal template", func() {
			BeforeEach(func() {
				// Empty selector matches every dataplane; no SpiffeID means the
				// default /workload/{{ .Workload }} path.
				Expect(builders.MeshIdentity().WithName("identity-1").Create(resStore)).To(Succeed())
			})

			It("should reject a tags-only token when the dataplane declares a workload label", func() {
				// when
				err := authenticate(&dpResWithWorkload, tagsOnlyToken())

				// then
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("requires a workload-bound token"))
			})

			It("should accept a workload-bound token and return its binding", func() {
				// when
				binding, err := authenticator.Authenticate(ctx, &dpResWithWorkload, generate(builtin_issuer.DataplaneIdentity{
					Mesh:     "default",
					Workload: "backend",
				}))

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(binding).To(Equal(auth.Binding{
					Labels: mesh_proto.MultiValueTagSet{
						metadata.KumaWorkload:  {"backend": true},
						metadata.KumaMeshLabel: {"default": true},
					},
				}))
			})

			It("should accept a token that binds the workload label through a tag", func() {
				// given
				dp := createDataplane("dp-workload-tag", map[string]string{metadata.KumaWorkload: "backend"}, metadata.KumaWorkload, "backend")

				// when
				binding, err := authenticator.Authenticate(ctx, dp, generate(builtin_issuer.DataplaneIdentity{
					Mesh: "default",
					Tags: map[string]map[string]bool{metadata.KumaWorkload: {"backend": true}},
				}))

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(binding.Labels).To(Equal(mesh_proto.MultiValueTagSet{
					metadata.KumaWorkload:  {"backend": true},
					metadata.KumaMeshLabel: {"default": true},
				}))
			})

			It("should prove only the workload both the claim and its tag allow", func() {
				// given
				dp := createDataplane("dp-workload-tag", map[string]string{metadata.KumaWorkload: "backend"}, metadata.KumaWorkload, "backend")

				// when
				binding, err := authenticator.Authenticate(ctx, dp, generate(builtin_issuer.DataplaneIdentity{
					Mesh:     "default",
					Workload: "backend",
					Tags:     map[string]map[string]bool{metadata.KumaWorkload: {"backend": true, "frontend": true}},
				}))

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(binding.Labels).To(Equal(mesh_proto.MultiValueTagSet{
					metadata.KumaWorkload:  {"backend": true},
					metadata.KumaMeshLabel: {"default": true},
				}))
			})

			It("should accept a tags-only token when the dataplane has no workload label", func() {
				Expect(authenticate(&dpRes, tagsOnlyToken())).To(Succeed())
			})
		})

		DescribeTable("should reject a tags-only token whenever the path reads the workload label",
			func(spiffeIDPath string) {
				// given
				createIdentity(spiffeIDPath)

				// when
				err := authenticate(&dpResWithWorkload, tagsOnlyToken())

				// then
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("requires a workload-bound token"))
			},
			Entry("canonical workload template", "/workload/{{ .Workload }}"),
			Entry("root-scoped .Workload", "/workload/{{ $.Workload }}"),
			Entry(".Workload with trim markers", "/workload/{{- .Workload -}}"),
			Entry("label with trim marker", "/workload/{{ label \"kuma.io/workload\" -}}"),
			Entry(".Workload through a pipeline", "/workload/{{ .Workload | printf \"%s\" }}"),
			Entry("workload read via a variable", "/workload/{{ $w := label \"kuma.io/workload\" }}{{ $w }}"),
			Entry("workload label via computed key", "/workload/{{ label (printf \"kuma.io/%s\" \"workload\") }}"),
			Entry("workload read on a branch of the trust domain", "{{ if eq .TrustDomain \"default.zone-1.mesh.local\" }}/workload/{{ .Workload }}{{ else }}/static{{ end }}"),
		)

		DescribeTable("should accept a tags-only token when the path reads no unbound label",
			func(spiffeIDPath string) {
				// given
				createIdentity(spiffeIDPath)

				// when / then
				Expect(authenticate(&dpResWithWorkload, tagsOnlyToken())).To(Succeed())
			},
			Entry("Kubernetes ns/sa template, labels absent on the dataplane", "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"),
			Entry("static path", "/custom/static"),
			Entry("label bound by the token", "/service/{{ label \"kuma.io/service\" }}"),
		)

		It("should require a tag for any other label the path reads", func() {
			// given a path built from the env label
			createIdentity("/env/{{ label \"env\" }}")
			dp := createDataplane("dp-env", map[string]string{"env": "prod"}, "env", "prod")

			// when the token doesn't bind env
			err := authenticate(dp, tagsOnlyToken())

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`built from its "env" label "prod", which the token does not bind`))

			// when the token binds env
			Expect(authenticate(dp, generate(builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Tags: map[string]map[string]bool{"kuma.io/service": {"web": true, "web-api": true}, "env": {"prod": true}},
			}))).To(Succeed())
		})

		It("should require a name-bound token when the path reads the display name", func() {
			// given a path built from the display name, which on Universal is the dataplane name
			createIdentity("/dp/{{ label \"kuma.io/display-name\" }}")
			dp := createDataplane("dp-named", map[string]string{mesh_proto.DisplayName: "dp-named"})

			// when the token doesn't bind the name
			err := authenticate(dp, tagsOnlyToken())

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("requires a name-bound token"))

			// when the token binds the name
			binding, err := authenticator.Authenticate(ctx, dp, generate(builtin_issuer.DataplaneIdentity{
				Mesh: "default",
				Name: "dp-named",
			}))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(binding.Labels).To(Equal(mesh_proto.MultiValueTagSet{
				mesh_proto.DisplayName: {"dp-named": true},
				metadata.KumaMeshLabel: {"default": true},
			}))
		})

		It("should bind the mesh label to the mesh the token is valid in", func() {
			// given a path built from the mesh label, which the dataplane may supply
			createIdentity("/mesh/{{ label \"kuma.io/mesh\" }}")
			inMesh := createDataplane("dp-in-mesh", map[string]string{metadata.KumaMeshLabel: "default"})
			otherMesh := createDataplane("dp-other-mesh", map[string]string{metadata.KumaMeshLabel: "demo"})

			// when / then
			Expect(authenticate(inMesh, tagsOnlyToken())).To(Succeed())
			err := authenticate(otherMesh, tagsOnlyToken())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`built from its "kuma.io/mesh" label "demo", which the token does not bind`))
		})

		It("should accept a tags-only token with a workload label when no MeshIdentity matches", func() {
			Expect(authenticate(&dpResWithWorkload, tagsOnlyToken())).To(Succeed())
		})
	})

	It("should throw an error on invalid token", func() {
		// when
		_, err := authenticator.Authenticate(ctx, &dpRes, "this-is-not-valid-jwt-token")

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
		_, err = authenticator.Authenticate(ctx, &dpRes, credential)

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
			_, err = authenticator.Authenticate(ctx, &ziRes, token)

			// then
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
