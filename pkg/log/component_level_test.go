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

	type override struct {
		component string
		level     kuma_log.LogLevel
	}

	DescribeTable("GetEffectiveLevelForNames",
		func(overrides []override, query string, expectFound bool, expectLevel kuma_log.LogLevel) {
			for _, o := range overrides {
				Expect(registry.SetLevel(o.component, o.level)).To(Succeed())
			}
			names := kuma_log.SplitHierarchy(query)
			if query == "" {
				names = nil
			}
			level, ok := registry.GetEffectiveLevelForNames(names)
			Expect(ok).To(Equal(expectFound))
			if expectFound {
				Expect(level).To(Equal(expectLevel))
			}
		},
		Entry("no overrides set", nil, "xds", false, kuma_log.LogLevel(0)),
		Entry("exact match",
			[]override{{"xds", kuma_log.DebugLevel}},
			"xds", true, kuma_log.DebugLevel),
		Entry("walks up hierarchy",
			[]override{{"xds", kuma_log.DebugLevel}},
			"xds.server", true, kuma_log.DebugLevel),
		Entry("prefers exact match over ancestor",
			[]override{{"xds", kuma_log.DebugLevel}, {"xds.server", kuma_log.InfoLevel}},
			"xds.server", true, kuma_log.InfoLevel),
		Entry("does not match unrelated components",
			[]override{{"xds", kuma_log.DebugLevel}},
			"kds", false, kuma_log.LogLevel(0)),
		Entry("deep hierarchy",
			[]override{{"plugins", kuma_log.DebugLevel}},
			"plugins.authn.api-server.tokens", true, kuma_log.DebugLevel),
		Entry("empty component name",
			[]override{{"xds", kuma_log.DebugLevel}},
			"", false, kuma_log.LogLevel(0)),
	)

	It("should reset single level", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
		registry.ResetLevel("xds")

		_, ok := registry.GetEffectiveLevelForNames(kuma_log.SplitHierarchy("xds"))
		Expect(ok).To(BeFalse())
	})

	It("should reset all levels", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
		Expect(registry.SetLevel("kds", kuma_log.InfoLevel)).To(Succeed())
		registry.ResetAll()

		_, ok := registry.GetEffectiveLevelForNames(kuma_log.SplitHierarchy("xds"))
		Expect(ok).To(BeFalse())
		_, ok = registry.GetEffectiveLevelForNames(kuma_log.SplitHierarchy("kds"))
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
			registry.GetEffectiveLevelForNames(kuma_log.SplitHierarchy("xds.server"))
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
		Expect(registry.SetLevel("component-0", kuma_log.InfoLevel)).To(Succeed())
	})
})
