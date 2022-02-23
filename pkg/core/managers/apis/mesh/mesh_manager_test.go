package mesh_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/managers/apis/mesh"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secrets_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	secrets_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/core/validators"
	ca_builtin "github.com/kumahq/kuma/pkg/plugins/ca/builtin"
	"github.com/kumahq/kuma/pkg/plugins/ca/provided"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_resources "github.com/kumahq/kuma/pkg/test/resources"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Mesh Manager", func() {

	var resManager manager.ResourceManager
	var secretManager manager.ResourceManager
	var resStore store.ResourceStore
	var builtinCaManager core_ca.Manager

	BeforeEach(func() {
		resStore = memory.NewStore()
		secretManager = secrets_manager.NewSecretManager(secrets_store.NewSecretStore(resStore), cipher.None(), nil)
		builtinCaManager = ca_builtin.NewBuiltinCaManager(secretManager)
		providedCaManager := provided.NewProvidedCaManager(datasource.NewDataSourceLoader(secretManager))
		caManagers := core_ca.Managers{
			"builtin":  builtinCaManager,
			"provided": providedCaManager,
		}

		manager := manager.NewResourceManager(resStore)
		validator := mesh.NewMeshValidator(caManagers, resStore)
		resManager = mesh.NewMeshManager(resStore, manager, caManagers, test_resources.Global(), validator)
	})

	Describe("Create()", func() {
		It("should also ensure that CAs are created", func() {
			// given
			meshName := "mesh-1"
			resKey := model.ResourceKey{
				Name: meshName,
			}

			// when
			mesh := core_mesh.MeshResource{
				Spec: &mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "builtin-1",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "builtin-1",
								Type: "builtin",
							},
						},
					},
				},
			}
			err := resManager.Create(context.Background(), &mesh, store.CreateBy(resKey))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and enabled CA is created
			_, err = builtinCaManager.GetRootCert(context.Background(), meshName, mesh.Spec.Mtls.Backends[0])
			Expect(err).ToNot(HaveOccurred())
		})

		It("should create default resources", func() {
			// given
			meshName := "mesh-1"
			resKey := model.ResourceKey{
				Name: meshName,
			}

			// when
			mesh := core_mesh.NewMeshResource()
			err := resManager.Create(context.Background(), mesh, store.CreateBy(resKey))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and default TrafficPermission for the mesh exists
			err = resStore.Get(context.Background(), core_mesh.NewTrafficPermissionResource(), store.GetByKey("allow-all-mesh-1", meshName))
			Expect(err).ToNot(HaveOccurred())

			// and default TrafficRoute for the mesh exists
			err = resStore.Get(context.Background(), core_mesh.NewTrafficRouteResource(), store.GetByKey("route-all-mesh-1", meshName))
			Expect(err).ToNot(HaveOccurred())

			// and Dataplane Token Signing Key for the mesh exists
			key := tokens.SigningKeyResourceKey(issuer.DataplaneTokenSigningKeyPrefix(meshName), tokens.DefaultSerialNumber, meshName)
			err = secretManager.Get(context.Background(), system.NewSecretResource(), store.GetBy(key))
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("should set default values for Prometheus settings", func() {

			type testCase struct {
				input    string
				expected string
			}

			DescribeTable("should apply defaults on a target MeshResource",
				func(given testCase) {
					// given
					key := model.ResourceKey{Name: "demo"}
					mesh := core_mesh.NewMeshResource()

					// when
					err := util_proto.FromYAML([]byte(given.input), mesh.Spec)
					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					err = resManager.Create(context.Background(), mesh, store.CreateBy(key))
					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					actual, err := util_proto.ToYAML(mesh.Spec)
					// then
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(MatchYAML(given.expected))

					By("fetching a fresh Mesh object")

					new := core_mesh.NewMeshResource()

					// when
					err = resManager.Get(context.Background(), new, store.GetBy(key))
					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					actual, err = util_proto.ToYAML(new.Spec)
					// then
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(MatchYAML(given.expected))
				},
				Entry("when both `metrics.prometheus.port` and `metrics.prometheus.path` are not set", testCase{
					input: `
                    metrics:
                      backends:
                      - name: prometheus-1
                        type: prometheus
`,
					expected: `
                    metrics:
                      backends:
                      - name: prometheus-1
                        type: prometheus
                        conf:
                          port: 5670
                          path: /metrics
                          tags:
                            kuma.io/service: dataplane-metrics
`,
				}),
			)
		})

		It("should validate all CAs", func() {
			// given
			meshName := "mesh-1"
			resKey := model.ResourceKey{
				Name: meshName,
			}

			// when
			mesh := core_mesh.MeshResource{
				Spec: &mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "ca-1",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "ca-1",
								Type: "provided",
							},
						},
					},
				},
			}
			err := resManager.Create(context.Background(), &mesh, store.CreateBy(resKey))

			// then
			Expect(err).To(MatchError("mtls.backends[0].conf.cert: has to be defined; mtls.backends[0].conf.key: has to be defined"))
		})
	})

	Describe("Delete()", func() {
		It("should delete secrets within one mesh", func() {
			// given two meshes
			err := resManager.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("demo-1", model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
			err = resManager.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("demo-2", model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			// when demo-1 is deleted
			err = resManager.Delete(context.Background(), core_mesh.NewMeshResource(), store.DeleteByKey("demo-1", model.NoMesh))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and all secrets are deleted
			secrets := &system.SecretResourceList{}
			err = secretManager.List(context.Background(), secrets, store.ListByMesh("demo-1"))
			Expect(err).ToNot(HaveOccurred())
			Expect(secrets.Items).To(BeEmpty())

			// and all secrets from other mesh are preserved
			secrets = &system.SecretResourceList{}
			err = secretManager.List(context.Background(), secrets, store.ListByMesh("demo-2"))
			Expect(err).ToNot(HaveOccurred())
			Expect(secrets.Items).To(HaveLen(1)) // default signing key
		})

		It("should not delete Mesh if there are Dataplanes attached", func() {
			// given mesh and dataplane
			err := resManager.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			err = resStore.Create(context.Background(), &core_mesh.DataplaneResource{
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
							Port:        8080,
							ServicePort: 80,
							Tags: map[string]string{
								"service": "mobile",
								"version": "v1",
							}},
						},
					},
				},
			}, store.CreateByKey("dp-1", "mesh-1"))
			Expect(err).ToNot(HaveOccurred())

			// when mesh-1 is delete
			err = resManager.Delete(context.Background(), core_mesh.NewMeshResource(), store.DeleteByKey("mesh-1", model.NoMesh))
			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("mesh: unable to delete mesh, there are still some dataplanes attached"))
		})
	})

	Describe("Update()", func() {
		It("should not allow to change CA when mTLS is enabled", func() {
			// given
			meshName := "mesh-1"
			resKey := model.ResourceKey{
				Name: meshName,
			}

			// when
			mesh := core_mesh.MeshResource{
				Spec: &mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "builtin-1",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "builtin-1",
								Type: "builtin",
							},
						},
					},
				},
			}
			err := resManager.Create(context.Background(), &mesh, store.CreateBy(resKey))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when trying to change CA
			mesh.Spec.Mtls.Backends[0].Name = "builtin-2"
			mesh.Spec.Mtls.EnabledBackend = "builtin-2"
			err = resManager.Update(context.Background(), &mesh)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   "mtls.enabledBackend",
						Message: "Changing CA when mTLS is enabled is forbidden. Disable mTLS first and then change the CA",
					},
				},
			}))
		})

		It("should allow to change CA when mTLS is disabled", func() {
			// given
			meshName := "mesh-1"
			resKey := model.ResourceKey{
				Name: meshName,
			}

			// when
			mesh := core_mesh.MeshResource{
				Spec: &mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "builtin-1",
								Type: "builtin",
							},
						},
					},
				},
			}
			err := resManager.Create(context.Background(), &mesh, store.CreateBy(resKey))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when trying to enable mTLS change CA
			mesh.Spec.Mtls.Backends[0].Name = "builtin-2"
			mesh.Spec.Mtls.EnabledBackend = "builtin-2"
			err = resManager.Update(context.Background(), &mesh)

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("should set default values for Prometheus settings", func() {

			type testCase struct {
				initial  string
				updated  string
				expected string
			}

			DescribeTable("should apply defaults on a target MeshResource",
				func(given testCase) {
					// given
					key := model.ResourceKey{Name: "demo"}
					mesh := core_mesh.NewMeshResource()

					// when
					err := util_proto.FromYAML([]byte(given.initial), mesh.Spec)
					// then
					Expect(err).ToNot(HaveOccurred())

					By("creating a new Mesh")
					// when
					err = resManager.Create(context.Background(), mesh, store.CreateBy(key))
					// then
					Expect(err).ToNot(HaveOccurred())

					By("changing Prometheus settings")
					// when
					mesh.Spec = &mesh_proto.Mesh{}
					err = util_proto.FromYAML([]byte(given.updated), mesh.Spec)
					// then
					Expect(err).ToNot(HaveOccurred())

					By("updating the Mesh with new Prometheus settings")
					// when
					err = resManager.Update(context.Background(), mesh)
					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					actual, err := util_proto.ToYAML(mesh.Spec)
					// then
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(MatchYAML(given.expected))

					By("fetching a fresh Mesh object")

					new := core_mesh.NewMeshResource()

					// when
					err = resManager.Get(context.Background(), new, store.GetBy(key))
					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					actual, err = util_proto.ToYAML(new.Spec)
					// then
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(MatchYAML(given.expected))
				},
				Entry("when both config is changed", testCase{
					initial: `
                    metrics:
                      enabledBackend: prometheus-1
                      backends:
                      - name: prometheus-1
                        type: prometheus
`,
					updated: `
                    metrics:
                      enabledBackend: prometheus-1
                      backends:
                      - name: prometheus-1
                        type: prometheus
                        conf:
                          port: 1234
                          path: /non-standard-path
                          tags:
                            kuma.io/service: custom-prom
`,
					expected: `
                    metrics:
                      enabledBackend: prometheus-1
                      backends:
                      - name: prometheus-1
                        type: prometheus
                        conf:
                          port: 1234
                          path: /non-standard-path
                          tags:
                            kuma.io/service: custom-prom
`,
				}),
				Entry("when config remain unchanged", testCase{
					initial: `
                    metrics:
                      enabledBackend: prometheus-1
                      backends:
                      - name: prometheus-1
                        type: prometheus
                        conf:
                          port: 1234
                          path: /non-standard-path
                          tags:
                            kuma.io/service: custom-prom
`,
					updated: `
                    metrics:
                      enabledBackend: prometheus-1
                      backends:
                      - name: prometheus-1
                        type: prometheus
                        conf:
                          port: 1234
                          path: /non-standard-path
                          tags:
                            kuma.io/service: custom-prom
`,
					expected: `
                    metrics:
                      enabledBackend: prometheus-1
                      backends:
                      - name: prometheus-1
                        type: prometheus
                        conf:
                          port: 1234
                          path: /non-standard-path
                          tags:
                            kuma.io/service: custom-prom
`,
				}),
			)
		})
	})
})
