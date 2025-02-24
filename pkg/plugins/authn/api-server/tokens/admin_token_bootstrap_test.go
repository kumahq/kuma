package tokens_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	store_config "github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_tokens "github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("Admin Token Bootstrap", func() {
	It("should bootstrap admin token", func() {
		// given
		ctx := context.Background()
		resManager := manager.NewResourceManager(memory.NewStore())
		signingKeyManager := core_tokens.NewSigningKeyManager(resManager, system.UserTokenSigningKeyPrefix)
		tokenIssuer := issuer.NewUserTokenIssuer(core_tokens.NewTokenIssuer(signingKeyManager))
		tokenValidator := issuer.NewUserTokenValidator(
			core_tokens.NewValidator(
				core.Log.WithName("test"),
				[]core_tokens.SigningKeyAccessor{
					core_tokens.NewSigningKeyAccessor(resManager, system.UserTokenSigningKeyPrefix),
				},
				core_tokens.NewRevocations(resManager, core_model.ResourceKey{Name: system.UserTokenRevocations}),
				store_config.MemoryStore,
			),
		)

		component := tokens.NewAdminTokenBootstrap(tokenIssuer, resManager, kuma_cp.DefaultConfig())
		err := signingKeyManager.CreateDefaultSigningKey(ctx)
		Expect(err).ToNot(HaveOccurred())
		stopCh := make(chan struct{})
		defer close(stopCh)

		// when
		go func() {
			_ = component.Start(stopCh) // it never returns an error
		}()

		// then token is created
		Eventually(func(g Gomega) {
			globalSecret := system.NewGlobalSecretResource()
			err = resManager.Get(context.Background(), globalSecret, core_store.GetBy(core_model.ResourceKey{Name: system.AdminUserToken}))
			g.Expect(err).ToNot(HaveOccurred())
			user, err := tokenValidator.Validate(ctx, string(globalSecret.Spec.Data.Value))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(user.Name).To(Equal("mesh-system:admin"))
			g.Expect(user.Groups).To(Equal([]string{"mesh-system:admin"}))
		}).Should(Succeed())
	})
})
