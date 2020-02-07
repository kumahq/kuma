package mesh

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/ca/builtin"
	builtin_issuer "github.com/Kong/kuma/pkg/core/ca/builtin/issuer"
	"github.com/Kong/kuma/pkg/core/ca/provided"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/secrets/cipher"
	secrets_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	secrets_store "github.com/Kong/kuma/pkg/core/secrets/store"
	"github.com/Kong/kuma/pkg/core/validators"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_resources "github.com/Kong/kuma/pkg/test/resources"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("Mesh Manager", func() {

	var resManager manager.ResourceManager
	var resStore store.ResourceStore
	var builtinCaManager builtin.BuiltinCaManager
	var providedCaManager provided.ProvidedCaManager

	BeforeEach(func() {
		resStore = memory.NewStore()
		secretManager := secrets_manager.NewSecretManager(secrets_store.NewSecretStore(resStore), cipher.None())
		builtinCaManager = builtin.NewBuiltinCaManager(secretManager)
		providedCaManager = provided.NewProvidedCaManager(secretManager)
		manager := manager.NewResourceManager(resStore)
		resManager = NewMeshManager(resStore, builtinCaManager, providedCaManager, manager, secretManager, test_resources.Global())
	})

	createProvidedCa := func(meshName string) string {
		// when
		signingPair, err := builtin_issuer.NewRootCA(meshName)
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		signingCert, err := providedCaManager.AddSigningCert(context.Background(), meshName, *signingPair)
		// then
		Expect(err).ToNot(HaveOccurred())

		return signingCert.Id
	}

	Describe("Create()", func() {
		It("should also create a built-in CA", func() {
			// given
			meshName := "mesh-1"
			resKey := model.ResourceKey{
				Mesh: meshName,
				Name: meshName,
			}

			// when
			mesh := core_mesh.MeshResource{}
			err := resManager.Create(context.Background(), &mesh, store.CreateBy(resKey))

			// then
			Expect(err).ToNot(HaveOccurred())

			// and built-in CA is created
			certs, err := builtinCaManager.GetRootCerts(context.Background(), meshName)
			Expect(err).ToNot(HaveOccurred())
			Expect(certs).To(HaveLen(1))
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
                    mtls:
                      ca:
                        builtin: {}
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
                    mtls:
                      ca:
                        builtin: {}
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
                    mtls:
                      ca:
                        builtin: {}
                    metrics:
                      prometheus:
                        port: 1234
                        path: /metrics
`,
				}),
			)
		})

		Context("Mesh with Provided CA", func() {
			// given
			meshName := "mesh-1"
			resKey := model.ResourceKey{
				Mesh: meshName,
				Name: meshName,
			}
			var mesh core_mesh.MeshResource

			BeforeEach(func() {
				mesh = core_mesh.MeshResource{
					Spec: mesh_proto.Mesh{
						Mtls: &mesh_proto.Mesh_Mtls{
							Enabled: true,
							Ca: &mesh_proto.CertificateAuthority{
								Type: &mesh_proto.CertificateAuthority_Provided_{
									Provided: &mesh_proto.CertificateAuthority_Provided{},
								},
							},
						},
					},
				}
			})

			It("should not allow provided CA when it's not created", func() {
				// when
				err := resManager.Create(context.Background(), &mesh, store.CreateBy(resKey))

				// then
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(&validators.ValidationError{
					Violations: []validators.Violation{
						{
							Field:   "mtls.ca.provided",
							Message: "There is no signing certificate in provided CA for a given mesh. Add certificate via 'kumactl manage ca provided certificates add' command.",
						},
					},
				}))
			})

			It("should not allow provided CA has no certs", func() {
				// given ca with no certs
				id := createProvidedCa("mesh-1")
				err := providedCaManager.DeleteSigningCert(context.Background(), meshName, id)
				Expect(err).ToNot(HaveOccurred())

				// when
				err = resManager.Create(context.Background(), &mesh, store.CreateBy(resKey))

				// then
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(&validators.ValidationError{
					Violations: []validators.Violation{
						{
							Field:   "mtls.ca.provided",
							Message: "There is no signing certificate in provided CA for a given mesh. Add certificate via 'kumactl manage ca provided certificates add' command.",
						},
					},
				}))
			})

			It("should allow provided CA when it's available", func() {
				// setup
				createProvidedCa("mesh-1")

				// when
				err := resManager.Create(context.Background(), &mesh, store.CreateBy(resKey))

				// then
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Update()", func() {
		It("should not allow to change CA when mTLS is enabled", func() {
			// setup

			// when
			signingPair, err := builtin_issuer.NewRootCA("mesh-1")
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			_, err = providedCaManager.AddSigningCert(context.Background(), "mesh-1", *signingPair)
			// then
			Expect(err).ToNot(HaveOccurred())

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
						Enabled: true,
						Ca: &mesh_proto.CertificateAuthority{
							Type: &mesh_proto.CertificateAuthority_Builtin_{
								Builtin: &mesh_proto.CertificateAuthority_Builtin{},
							},
						},
					},
				},
			}
			err = resManager.Create(context.Background(), &mesh, store.CreateBy(resKey))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when trying to change CA
			mesh.Spec.Mtls.Ca = &mesh_proto.CertificateAuthority{
				Type: &mesh_proto.CertificateAuthority_Provided_{
					Provided: &mesh_proto.CertificateAuthority_Provided{},
				},
			}
			err = resManager.Update(context.Background(), &mesh)

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   "mtls.ca",
						Message: "Changing CA when mTLS is enabled is forbidden. Disable mTLS first and then change the CA",
					},
				},
			}))
		})

		It("should allow to change CA when mTLS is disabled", func() {
			// setup
			createProvidedCa("mesh-1")

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
						Enabled: false,
						Ca: &mesh_proto.CertificateAuthority{
							Type: &mesh_proto.CertificateAuthority_Builtin_{
								Builtin: &mesh_proto.CertificateAuthority_Builtin{},
							},
						},
					},
				},
			}
			err := resManager.Create(context.Background(), &mesh, store.CreateBy(resKey))

			// then
			Expect(err).ToNot(HaveOccurred())

			// when trying to enable mTLS change CA
			mesh.Spec.Mtls = &mesh_proto.Mesh_Mtls{
				Enabled: true,
				Ca: &mesh_proto.CertificateAuthority{
					Type: &mesh_proto.CertificateAuthority_Provided_{
						Provided: &mesh_proto.CertificateAuthority_Provided{},
					},
				},
			}
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
                    mtls:
                      ca:
                        builtin: {}
                    metrics: {}
`,
				}),
				Entry("when `metrics` is unset", testCase{
					initial: `
                    metrics:
                      prometheus: {}
`,
					updated: ``,
					expected: `
                    mtls:
                      ca:
                        builtin: {}
`,
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
                    mtls:
                      ca:
                        builtin: {}
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
                    mtls:
                      ca:
                        builtin: {}
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
                    mtls:
                      ca:
                        builtin: {}
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
                    mtls:
                      ca:
                        builtin: {}
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
                    mtls:
                      ca:
                        builtin: {}
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
			// setup
			createProvidedCa("mesh-1")

			// given mesh
			meshName := "mesh-1"

			mesh := core_mesh.MeshResource{}
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
			_, err = builtinCaManager.GetRootCerts(context.Background(), meshName)
			Expect(err).ToNot(BeNil())
			Expect(err).To(MatchError("failed to load CA key pair for Mesh \"mesh-1\": Resource not found: type=\"Secret\" name=\"builtinca.mesh-1\" mesh=\"mesh-1\"")) // todo(jakubdyszkiewicz) make error msg consistent

			// and provided mesh CA is deleted
			_, err = providedCaManager.GetSigningCerts(context.Background(), meshName)
			Expect(err).ToNot(BeNil())
			Expect(store.IsResourceNotFound(err)).To(BeTrue())
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
