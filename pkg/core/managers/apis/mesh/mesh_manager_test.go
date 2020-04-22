package mesh

import (
	"context"

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
	var builtinCaManager core_ca.CaManager

	BeforeEach(func() {
		resStore = memory.NewStore()
		secretManager := secrets_manager.NewSecretManager(secrets_store.NewSecretStore(resStore), cipher.None())
		builtinCaManager = ca_builtin.NewBuiltinCaManager(secretManager)
		caManagers := core_ca.CaManagers{
			"builtin": builtinCaManager,
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
						DefaultBackend: "builtin-1",
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

			// and defined CAs are created
			_, err = builtinCaManager.GetRootCert(context.Background(), meshName, *mesh.Spec.Mtls.Backends[0])
			Expect(err).ToNot(HaveOccurred())
			_, err = builtinCaManager.GetRootCert(context.Background(), meshName, *mesh.Spec.Mtls.Backends[1])
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
                      prometheus: {}
`,
					expected: `
                    metrics:
                      prometheus:
                        port: 5670
                        path: /metrics
`,
				}),
				Entry("when `metrics.prometheus.port` is not set", testCase{
					input: `
                    metrics:
                      prometheus:
                        path: /non-standard-path
`,
					expected: `
                    metrics:
                      prometheus:
                        port: 5670
                        path: /non-standard-path
`,
				}),
				Entry("when `metrics.prometheus.path` is not set", testCase{
					input: `
                    metrics:
                      prometheus:
                        port: 1234
`,
					expected: `
                    metrics:
                      prometheus:
                        port: 1234
                        path: /metrics
`,
				}),
			)
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
						Enabled:        true,
						DefaultBackend: "builtin-1",
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
			mesh.Spec.Mtls.DefaultBackend = "builtin-2"
			err = resManager.Update(context.Background(), &mesh)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   "mtls.defaultBackend",
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
						Enabled:        false,
						DefaultBackend: "builtin-1",
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
			mesh.Spec.Mtls.DefaultBackend = "builtin-2"
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
				Entry("when `metrics.prometheus` is unset", testCase{
					initial: `
                    metrics:
                      prometheus: {}
`,
					updated: `
                    metrics: {}
`,
					expected: `
                    metrics: {}
`,
				}),
				Entry("when `metrics` is unset", testCase{
					initial: `
                    metrics:
                      prometheus: {}
`,
					updated:  ``,
					expected: `{}`,
				}),
				Entry("when both `metrics.prometheus.port` and `metrics.prometheus.path` are unset", testCase{
					initial: `
                    metrics:
                      prometheus: {}
`,
					updated: `
                    metrics:
                      prometheus: {}
`,
					expected: `
                    metrics:
                      prometheus:
                        port: 5670
                        path: /metrics
`,
				}),
				Entry("when `metrics.prometheus.port` is unset", testCase{
					initial: `
                    metrics:
                      prometheus: {}
`,
					updated: `
                    metrics:
                      prometheus:
                        path: /non-standard-path
`,
					expected: `
                    metrics:
                      prometheus:
                        port: 5670
                        path: /non-standard-path
`,
				}),
				Entry("when `metrics.prometheus.path` is unset", testCase{
					initial: `
                    metrics:
                      prometheus: {}
`,
					updated: `
                    metrics:
                      prometheus:
                        port: 1234
`,
					expected: `
                    metrics:
                      prometheus:
                        port: 1234
                        path: /metrics
`,
				}),
				Entry("when both `metrics.prometheus.port` and `metrics.prometheus.path` are changed", testCase{
					initial: `
                    metrics:
                      prometheus: {}
`,
					updated: `
                    metrics:
                      prometheus:
                        port: 1234
                        path: /non-standard-path
`,
					expected: `
                    metrics:
                      prometheus:
                        port: 1234
                        path: /non-standard-path
`,
				}),
				Entry("when both `metrics.prometheus.port` and `metrics.prometheus.path` remain unchanged", testCase{
					initial: `
                    metrics:
                      prometheus:
                        port: 1234
                        path: /non-standard-path
`,
					updated: `
                    metrics:
                      prometheus:
                        port: 1234
                        path: /non-standard-path
`,
					expected: `
                    metrics:
                      prometheus:
                        port: 1234
                        path: /non-standard-path
`,
				}),
			)
		})
	})

	Describe("Delete()", func() {
		It("should delete all associated resources", func() {
			// given mesh
			meshName := "mesh-1"

			mesh := core_mesh.MeshResource{
				Spec: mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						DefaultBackend: "builtin-1",
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
			resKey := model.ResourceKey{
				Mesh: meshName,
				Name: meshName,
			}
			err := resManager.Create(context.Background(), &mesh, store.CreateBy(resKey))
			Expect(err).ToNot(HaveOccurred())

			// and resource associated with it
			dp := core_mesh.DataplaneResource{}
			err = resStore.Create(context.Background(), &dp, store.CreateByKey("dp-1", meshName))
			Expect(err).ToNot(HaveOccurred())

			// when mesh is deleted
			err = resManager.Delete(context.Background(), &mesh, store.DeleteBy(resKey))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and resource is deleted
			err = resStore.Get(context.Background(), &core_mesh.DataplaneResource{}, store.GetByKey("dp-1", meshName))
			Expect(store.IsResourceNotFound(err)).To(BeTrue())

			// and built-in mesh CA is deleted
			_, err = builtinCaManager.GetRootCert(context.Background(), meshName, *mesh.Spec.Mtls.Backends[0])
			Expect(err).ToNot(BeNil())
			Expect(err).To(MatchError(`failed to load CA key pair for Mesh "mesh-1" and backend "builtin-1": Resource not found: type="Secret" name="ca-builtin-cert-builtin-1" mesh="mesh-1"`)) // todo(jakubdyszkiewicz) make error msg consistent
		})

		It("should delete all associated resources even if mesh is already removed", func() {
			// given resource that was not deleted with mesh
			dp := core_mesh.DataplaneResource{}
			dpKey := model.ResourceKey{
				Mesh: "already-deleted",
				Name: "dp-1",
			}
			err := resStore.Create(context.Background(), &dp, store.CreateBy(dpKey))
			Expect(err).ToNot(HaveOccurred())

			// when
			mesh := core_mesh.MeshResource{}
			err = resManager.Delete(context.Background(), &mesh, store.DeleteByKey("already-deleted", "already-deleted"))

			// then not found error is thrown
			Expect(err).To(HaveOccurred())
			Expect(store.IsResourceNotFound(err)).To(BeTrue())

			// but the resource in this mesh is deleted anyway
			err = resStore.Get(context.Background(), &dp, store.GetBy(dpKey))
			Expect(store.IsResourceNotFound(err)).To(BeTrue())
		})
	})
})
