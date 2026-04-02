package log_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_log "github.com/kumahq/kuma/v2/pkg/log"
)

var _ = Describe("ComponentLevelRegistry", func() {
	var registry *kuma_log.ComponentLevelRegistry

	BeforeEach(func() {
		registry = kuma_log.NewComponentLevelRegistry()
	})

	It("should return false when no overrides set", func() {
		_, ok := registry.GetEffectiveLevel("xds")
		Expect(ok).To(BeFalse())
	})

	It("should return exact match", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())

		level, ok := registry.GetEffectiveLevel("xds")
		Expect(ok).To(BeTrue())
		Expect(level).To(Equal(kuma_log.DebugLevel))
	})

	It("should walk up hierarchy", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())

		level, ok := registry.GetEffectiveLevel("xds.server")
		Expect(ok).To(BeTrue())
		Expect(level).To(Equal(kuma_log.DebugLevel))
	})

	It("should prefer exact match over ancestor", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
		Expect(registry.SetLevel("xds.server", kuma_log.InfoLevel)).To(Succeed())

		level, ok := registry.GetEffectiveLevel("xds.server")
		Expect(ok).To(BeTrue())
		Expect(level).To(Equal(kuma_log.InfoLevel))
	})

	It("should not match unrelated components", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())

		_, ok := registry.GetEffectiveLevel("kds")
		Expect(ok).To(BeFalse())
	})

	It("should handle deep hierarchy", func() {
		Expect(registry.SetLevel("plugins", kuma_log.DebugLevel)).To(Succeed())

		level, ok := registry.GetEffectiveLevel("plugins.authn.api-server.tokens")
		Expect(ok).To(BeTrue())
		Expect(level).To(Equal(kuma_log.DebugLevel))
	})

	It("should reset single level", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
		registry.ResetLevel("xds")

		_, ok := registry.GetEffectiveLevel("xds")
		Expect(ok).To(BeFalse())
	})

	It("should reset all levels", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
		Expect(registry.SetLevel("kds", kuma_log.InfoLevel)).To(Succeed())
		registry.ResetAll()

		_, ok := registry.GetEffectiveLevel("xds")
		Expect(ok).To(BeFalse())
		_, ok = registry.GetEffectiveLevel("kds")
		Expect(ok).To(BeFalse())
	})

	It("should list overrides", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
		Expect(registry.SetLevel("kds", kuma_log.InfoLevel)).To(Succeed())

		overrides := registry.ListOverrides()
		Expect(overrides).To(HaveLen(2))
		Expect(overrides["xds"]).To(Equal(kuma_log.DebugLevel))
		Expect(overrides["kds"]).To(Equal(kuma_log.InfoLevel))
	})

	It("should return false for empty component name", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())

		_, ok := registry.GetEffectiveLevel("")
		Expect(ok).To(BeFalse())
	})

	It("should handle concurrent reads and writes", func() {
		done := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(done)
			for range 1000 {
				_ = registry.SetLevel("xds", kuma_log.DebugLevel)
				_ = registry.SetLevel("kds", kuma_log.InfoLevel)
				registry.ResetLevel("xds")
				registry.ResetAll()
			}
		}()
		for range 1000 {
			registry.GetEffectiveLevel("xds.server")
			registry.ListOverrides()
		}
		<-done
	})

	It("should reject when max overrides exceeded", func() {
		for i := range kuma_log.MaxOverrides {
			Expect(registry.SetLevel(fmt.Sprintf("component-%d", i), kuma_log.DebugLevel)).To(Succeed())
		}
		err := registry.SetLevel("one-too-many", kuma_log.DebugLevel)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("maximum"))
	})

	It("should allow updating existing override when at max", func() {
		for i := range kuma_log.MaxOverrides {
			Expect(registry.SetLevel(fmt.Sprintf("component-%d", i), kuma_log.DebugLevel)).To(Succeed())
		}
		// Updating existing key should succeed
		Expect(registry.SetLevel("component-0", kuma_log.InfoLevel)).To(Succeed())
	})
})
