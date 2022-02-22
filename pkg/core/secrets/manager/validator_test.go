package manager_test

import (
	"context"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_datasource "github.com/kumahq/kuma/pkg/core/datasource"
	core_managers "github.com/kumahq/kuma/pkg/core/managers/apis/mesh"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secrets_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	secrets_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	ca_builtin "github.com/kumahq/kuma/pkg/plugins/ca/builtin"
	ca_provided "github.com/kumahq/kuma/pkg/plugins/ca/provided"
	provided_config "github.com/kumahq/kuma/pkg/plugins/ca/provided/config"
	resources_memory "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Secret Validator", func() {

	var validator secrets_manager.SecretValidator
	var resManager core_manager.ResourceManager
	var caManagers core_ca.Managers

	BeforeEach(func() {
		memoryStore := resources_memory.NewStore()
		resManager = core_manager.NewResourceManager(memoryStore)
		caManagers = core_ca.Managers{}
		secrets_manager.NewSecretValidator(caManagers, memoryStore)
		validator = secrets_manager.NewSecretValidator(caManagers, memoryStore)
		secManager := secrets_manager.NewSecretManager(secrets_store.NewSecretStore(memoryStore), cipher.None(), validator)

		caManagers["builtin"] = ca_builtin.NewBuiltinCaManager(secManager)
		caManagers["provided"] = ca_provided.NewProvidedCaManager(core_datasource.NewDataSourceLoader(secManager))
	})

	type testCase struct {
		mesh       *core_mesh.MeshResource
		secretName string
		expected   string
	}
	DescribeTable("should validate that secrets are in use",
		func(given testCase) {
			// given
			err := resManager.Create(context.Background(), given.mesh, core_store.CreateByKey(model.DefaultMesh, model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
			err = core_managers.EnsureCAs(context.Background(), caManagers, given.mesh, model.DefaultMesh)
			Expect(err).ToNot(HaveOccurred())

			// when
			verr := validator.ValidateDelete(context.Background(), given.secretName, "default")

			// then
			actual, err := yaml.Marshal(verr)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("when secret is used in builtin CA", testCase{
			mesh: &core_mesh.MeshResource{
				Spec: &mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "ca-1",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "ca-1",
								Type: "builtin",
							},
						},
					},
				},
			},
			secretName: "default.ca-builtin-cert-ca-1",
			expected: `
            violations:
            - field: name
              message: The secret "default.ca-builtin-cert-ca-1" that you are trying to remove is currently in use in Mesh "default" in mTLS backend "ca-1". Please remove the reference from the "ca-1" backend before removing the secret.`,
		}),
		Entry("when secret is used in provided CA", testCase{
			mesh: &core_mesh.MeshResource{
				Spec: &mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "ca-2",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "ca-2",
								Type: "provided",
								Conf: proto.MustToStruct(&provided_config.ProvidedCertificateAuthorityConfig{
									Cert: &system_proto.DataSource{
										Type: &system_proto.DataSource_Secret{
											Secret: "my-ca-cert",
										},
									},
									Key: &system_proto.DataSource{
										Type: &system_proto.DataSource_Secret{
											Secret: "my-ca-key",
										},
									},
								}),
							},
						},
					},
				},
			},
			secretName: "my-ca-cert",
			expected: `
            violations:
            - field: name
              message: The secret "my-ca-cert" that you are trying to remove is currently in use in Mesh "default" in mTLS backend "ca-2". Please remove the reference from the "ca-2" backend before removing the secret.`,
		}),
	)

	It("should pass validation of secrets that are not in use", func() {
		// given
		err := resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// when
		err = validator.ValidateDelete(context.Background(), "some-not-used-secret", "default")

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should pass validation of secrets in mesh that non exist", func() {
		// when
		err := validator.ValidateDelete(context.Background(), "some-not-used-secret", "non-existing-mesh")

		// then
		Expect(err).ToNot(HaveOccurred())
	})
})
