package bootstrap

import (
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	builtin_issuer "github.com/kumahq/kuma/pkg/tokens/builtin/issuer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bootstrap", func() {

	It("should create a default signing key", func() {
		// given
		cfg := kuma_cp.DefaultConfig()
		rt, err := Bootstrap(cfg)
		Expect(err).ToNot(HaveOccurred())

		// when
		key, err := builtin_issuer.GetSigningKey(rt.ResourceManager())

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(key).ToNot(HaveLen(0))

		// when kuma-cp is run again
		err = onStartup(rt)
		Expect(err).ToNot(HaveOccurred())
		key2, err := builtin_issuer.GetSigningKey(rt.ResourceManager())

		// then it should skip creating a new signing key
		Expect(err).ToNot(HaveOccurred())
		Expect(key).To(Equal(key2))
	})
})
