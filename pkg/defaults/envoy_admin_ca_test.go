package defaults_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/defaults"
	envoy_admin_tls "github.com/kumahq/kuma/pkg/envoy/admin/tls"
	resources_memory "github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("Envoy Admin CA defaults", func() {
	It("should create Envoy Admin CA", func() {
		// given
		store := resources_memory.NewStore()
		manager := core_manager.NewResourceManager(store)
		component := defaults.EnvoyAdminCaDefaultComponent{
			ResManager: manager,
			Extensions: context.Background(),
		}

		// when
		err := component.Start(nil)

		// then
		Expect(err).ToNot(HaveOccurred())
		err = manager.Get(context.Background(), core_system.NewGlobalSecretResource(), core_store.GetBy(envoy_admin_tls.GlobalSecretKey))
		Expect(err).ToNot(HaveOccurred())
	})
})
