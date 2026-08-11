package manager_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	core_manager "github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/secrets/cipher"
	secrets_manager "github.com/kumahq/kuma/v3/pkg/core/secrets/manager"
	secrets_store "github.com/kumahq/kuma/v3/pkg/core/secrets/store"
	resources_memory "github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

var _ = Describe("Secret Manager", func() {
	var resManager core_manager.ResourceManager
	var secManager core_manager.ResourceManager

	BeforeEach(func() {
		memoryStore := resources_memory.NewStore()
		resManager = core_manager.NewResourceManager(memoryStore)
		secManager = secrets_manager.NewSecretManager(secrets_store.NewSecretStore(memoryStore), cipher.None())
	})

	It("should delete a Secret still referenced by a legacy mTLS backend", func() {
		// given a Secret
		secret := system.NewSecretResource()
		secret.Spec = &system_proto.Secret{
			Data: util_proto.Bytes([]byte("secret-value")),
		}
		Expect(secManager.Create(context.Background(), secret, core_store.CreateByKey("ca-1-cert", model.DefaultMesh))).To(Succeed())

		// and a Mesh whose mTLS backend references it
		mesh := core_mesh.NewMeshResource()
		mesh.Spec = &mesh_proto.Mesh{
			Mtls: &mesh_proto.Mesh_Mtls{
				EnabledBackend: "ca-1",
				Backends: []*mesh_proto.CertificateAuthorityBackend{{
					Name: "ca-1",
					Type: "provided",
					Conf: util_proto.MustToStruct(&system_proto.DataSource{
						Type: &system_proto.DataSource_Secret{Secret: "ca-1-cert"},
					}),
				}},
			},
		}
		Expect(resManager.Create(context.Background(), mesh, core_store.CreateByKey(model.DefaultMesh, model.NoMesh))).To(Succeed())

		// when the Secret is deleted
		err := secManager.Delete(context.Background(), system.NewSecretResource(), core_store.DeleteByKey("ca-1-cert", model.DefaultMesh))

		// then
		Expect(err).ToNot(HaveOccurred())
	})
})
