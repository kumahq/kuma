package mesh

import (
	"context"
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/ca/builtin"
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
	"github.com/Kong/kuma/pkg/tls"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		pair, err := tls.NewSelfSignedCert("kuma", tls.ServerCertType)
		Expect(err).ToNot(HaveOccurred())
		signingCert, err := providedCaManager.AddSigningCert(context.Background(), meshName, pair)
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
							Message: "There are no provided CA for a given mesh",
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
							Message: "There are no signing certificate in provided CA for a given mesh",
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
			pair, err := tls.NewSelfSignedCert("kuma", tls.ServerCertType)
			Expect(err).ToNot(HaveOccurred())
			_, err = providedCaManager.AddSigningCert(context.Background(), "mesh-1", pair)
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
			Expect(err).To(MatchError("failed to load CA key pair for Mesh \"mesh-1\": Resource not found: type=\"Secret\" name=\"builtinca.mesh-1\" mesh=\"mesh-1\""))

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
