package defaults_test

import (
	"context"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/defaults"
	resources_memory "github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("Envoy Admin CA defaults", func() {
	It("should create Envoy Admin CA", func() {
		// given
		store := resources_memory.NewStore()
		manager := core_manager.NewResourceManager(store)

		// when
		err := defaults.EnsureEnvoyAdminCaExists(context.Background(), manager, logr.Discard(), kuma_cp.Config{})

		// then
		Expect(err).ToNot(HaveOccurred())
		err = manager.Get(context.Background(), core_system.NewGlobalSecretResource(), core_store.GetBy(core_model.ResourceKey{Name: core_system.EnvoyAdminCA}))
		Expect(err).ToNot(HaveOccurred())
	})
})
