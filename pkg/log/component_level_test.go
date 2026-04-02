package log_test

import (
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
		registry.SetLevel("xds", kuma_log.DebugLevel)

		level, ok := registry.GetEffectiveLevel("xds")
		Expect(ok).To(BeTrue())
		Expect(level).To(Equal(kuma_log.DebugLevel))
	})

	It("should walk up hierarchy", func() {
		registry.SetLevel("xds", kuma_log.DebugLevel)

		level, ok := registry.GetEffectiveLevel("xds.server")
		Expect(ok).To(BeTrue())
		Expect(level).To(Equal(kuma_log.DebugLevel))
	})

	It("should prefer exact match over ancestor", func() {
		registry.SetLevel("xds", kuma_log.DebugLevel)
		registry.SetLevel("xds.server", kuma_log.InfoLevel)

		level, ok := registry.GetEffectiveLevel("xds.server")
		Expect(ok).To(BeTrue())
		Expect(level).To(Equal(kuma_log.InfoLevel))
	})

	It("should not match unrelated components", func() {
		registry.SetLevel("xds", kuma_log.DebugLevel)

		_, ok := registry.GetEffectiveLevel("kds")
		Expect(ok).To(BeFalse())
	})

	It("should handle deep hierarchy", func() {
		registry.SetLevel("plugins", kuma_log.DebugLevel)

		level, ok := registry.GetEffectiveLevel("plugins.authn.api-server.tokens")
		Expect(ok).To(BeTrue())
		Expect(level).To(Equal(kuma_log.DebugLevel))
	})

	It("should reset single level", func() {
		registry.SetLevel("xds", kuma_log.DebugLevel)
		registry.ResetLevel("xds")

		_, ok := registry.GetEffectiveLevel("xds")
		Expect(ok).To(BeFalse())
	})

	It("should reset all levels", func() {
		registry.SetLevel("xds", kuma_log.DebugLevel)
		registry.SetLevel("kds", kuma_log.InfoLevel)
		registry.ResetAll()

		_, ok := registry.GetEffectiveLevel("xds")
		Expect(ok).To(BeFalse())
		_, ok = registry.GetEffectiveLevel("kds")
		Expect(ok).To(BeFalse())
	})

	It("should list overrides", func() {
		registry.SetLevel("xds", kuma_log.DebugLevel)
		registry.SetLevel("kds", kuma_log.InfoLevel)

		overrides := registry.ListOverrides()
		Expect(overrides).To(HaveLen(2))
		Expect(overrides["xds"]).To(Equal(kuma_log.DebugLevel))
		Expect(overrides["kds"]).To(Equal(kuma_log.InfoLevel))
	})

	It("should return false for empty component name", func() {
		registry.SetLevel("xds", kuma_log.DebugLevel)

		_, ok := registry.GetEffectiveLevel("")
		Expect(ok).To(BeFalse())
	})
})
