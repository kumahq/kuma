package manager_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
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
	var secManager core_manager.ResourceManager

	BeforeEach(func() {
		memoryStore := resources_memory.NewStore()
		secManager = secrets_manager.NewSecretManager(secrets_store.NewSecretStore(memoryStore), cipher.None())
	})

	It("should delete a Secret without consulting any usage validator", func() {
		// given
		secret := system.NewSecretResource()
		secret.Spec = &system_proto.Secret{
			Data: util_proto.Bytes([]byte("secret-value")),
		}
		Expect(secManager.Create(context.Background(), secret, core_store.CreateByKey("ca-1-cert", model.DefaultMesh))).To(Succeed())

		// when
		err := secManager.Delete(context.Background(), system.NewSecretResource(), core_store.DeleteByKey("ca-1-cert", model.DefaultMesh))

		// then
		Expect(err).ToNot(HaveOccurred())
	})
})
