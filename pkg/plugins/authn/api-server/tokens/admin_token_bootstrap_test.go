package tokens_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens"
	"github.com/kumahq/kuma/pkg/plugins/authn/api-server/tokens/issuer"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("Admin Token Bootstrap", func() {
	It("should bootstrap admin token", func() {
		// given
		resManager := manager.NewResourceManager(memory.NewStore())
		signingKeyManager := issuer.NewSigningKeyManager(resManager)
		tokenIssuer := issuer.NewUserTokenIssuer(signingKeyManager, issuer.NewTokenRevocations(resManager))
		component := tokens.NewAdminTokenBootstrap(tokenIssuer, resManager, kuma_cp.DefaultConfig())
		err := signingKeyManager.CreateDefaultSigningKey()
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
			err = resManager.Get(context.Background(), globalSecret, core_store.GetBy(tokens.AdminTokenKey))
			g.Expect(err).ToNot(HaveOccurred())
			user, err := tokenIssuer.Validate(string(globalSecret.Spec.Data.Value))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(user.Name).To(Equal("admin"))
			g.Expect(user.Groups).To(Equal([]string{"admin"}))
		}).Should(Succeed())
	})
})
