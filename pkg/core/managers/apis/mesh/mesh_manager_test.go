package mesh

import (
	"context"

	"github.com/Kong/kuma/pkg/core/datasource"
	"github.com/Kong/kuma/pkg/plugins/ca/provided"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_ca "github.com/Kong/kuma/pkg/core/ca"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/secrets/cipher"
	secrets_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	secrets_store "github.com/Kong/kuma/pkg/core/secrets/store"
	"github.com/Kong/kuma/pkg/core/validators"
	ca_builtin "github.com/Kong/kuma/pkg/plugins/ca/builtin"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_resources "github.com/Kong/kuma/pkg/test/resources"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("Mesh Manager", func() {

	var resManager manager.ResourceManager
	var resStore store.ResourceStore
	var builtinCaManager core_ca.Manager

	BeforeEach(func() {
		resStore = memory.NewStore()
		secretManager := secrets_manager.NewSecretManager(secrets_store.NewSecretStore(resStore), cipher.None(), nil)
		builtinCaManager = ca_builtin.NewBuiltinCaManager(secretManager)
		providedCaManager := provided.NewProvidedCaManager(datasource.NewDataSourceLoader(secretManager))
		caManagers := core_ca.Managers{
			"builtin":  builtinCaManager,
			"provided": providedCaManager,
		}

		manager := manager.NewResourceManager(resStore)
		validator := MeshValidator{CaManagers: caManagers}
		resManager = NewMeshManager(resStore, manager, secretManager, caManagers, test_resources.Global(), validator)
	})

	Describe("Create()", func() {
		It("should also ensure that CAs are created", func() {
			// given
			meshName := "mesh-1"
			resKey := model.ResourceKey{
				Mesh: meshName,
				Name: meshName,
			}

			// when
			mesh := core_mesh.MeshResource{
				Spec: mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "builtin-1",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "builtin-1",
								Type: "builtin",
							},
							{
								Name: "builtin-2",
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
			_, err = builtinCaManager.GetRootCert(context.Background(), meshName, *mesh.Spec.Mtls.Backends[0])
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
					key := model.ResourceKey{Mesh: "demo", Name: "demo"}
					mesh := core_mesh.MeshResource{}

					// when
					err := util_proto.FromYAML([]byte(given.input), &mesh.Spec)
					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					err = resManager.Create(context.Background(), &mesh, store.CreateBy(key))
					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					actual, err := util_proto.ToYAML(&mesh.Spec)
					// then
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(MatchYAML(given.expected))

					By("fetching a fresh Mesh object")

					new := core_mesh.MeshResource{}

					// when
					err = resManager.Get(context.Background(), &new, store.GetBy(key))
					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					actual, err = util_proto.ToYAML(&new.Spec)
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
                            service: dataplane-metrics
`,
				}),
			)
		})

		It("should validate all CAs", func() {
			// given
			meshName := "mesh-1"
			resKey := model.ResourceKey{
				Mesh: meshName,
				Name: meshName,
			}

			// when
			mesh := core_mesh.MeshResource{
				Spec: mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "ca-1",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "ca-1",
								Type: "provided",
							},
							{
								Name: "ca-2",
								Type: "provided",
							},
						},
					},
				},
			}
			err := resManager.Create(context.Background(), &mesh, store.CreateBy(resKey))

			// then
			Expect(err).To(MatchError("mtls.backends[0].config.cert: has to be defined; mtls.backends[0].config.key: has to be defined; mtls.backends[1].config.cert: has to be defined; mtls.backends[1].config.key: has to be defined"))
		})
	})

	Describe("Update()", func() {
		It("should not allow to change CA when mTLS is enabled", func() {
			// given
			meshName := "mesh-1"
			resKey := model.ResourceKey{
				Mesh: meshName,
				Name: meshName,
			}

			// when
			mesh := core_mesh.MeshResource{
				Spec: mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "builtin-1",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "builtin-1",
								Type: "builtin",
							},
							{
								Name: "builtin-2",
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
				Mesh: meshName,
				Name: meshName,
			}

			// when
			mesh := core_mesh.MeshResource{
				Spec: mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "builtin-1",
								Type: "builtin",
							},
							{
								Name: "builtin-2",
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
					key := model.ResourceKey{Mesh: "demo", Name: "demo"}
					mesh := core_mesh.MeshResource{}

					// when
					err := util_proto.FromYAML([]byte(given.initial), &mesh.Spec)
					// then
					Expect(err).ToNot(HaveOccurred())

					By("creating a new Mesh")
					// when
					err = resManager.Create(context.Background(), &mesh, store.CreateBy(key))
					// then
					Expect(err).ToNot(HaveOccurred())

					By("changing Prometheus settings")
					// when
					mesh.Spec = mesh_proto.Mesh{}
					err = util_proto.FromYAML([]byte(given.updated), &mesh.Spec)
					// then
					Expect(err).ToNot(HaveOccurred())

					By("updating the Mesh with new Prometheus settings")
					// when
					err = resManager.Update(context.Background(), &mesh)
					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					actual, err := util_proto.ToYAML(&mesh.Spec)
					// then
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(MatchYAML(given.expected))

					By("fetching a fresh Mesh object")

					new := core_mesh.MeshResource{}

					// when
					err = resManager.Get(context.Background(), &new, store.GetBy(key))
					// then
					Expect(err).ToNot(HaveOccurred())

					// when
					actual, err = util_proto.ToYAML(&new.Spec)
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
                            service: custom-prom
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
                            service: custom-prom
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
                            service: custom-prom
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
                            service: custom-prom
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
                            service: custom-prom
`,
				}),
			)
		})
	})
})
