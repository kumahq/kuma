package log_test

import (
	"fmt"
	"strings"

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

	It("SetLevel with same value is a no-op", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
		Expect(registry.ListOverrides()).To(HaveLen(1))
	})

	It("ResetLevel of nonexistent component is idempotent", func() {
		registry.ResetLevel("nonexistent")
		Expect(registry.ListOverrides()).To(BeEmpty())
	})

	It("ResetAll returns exact overrides that were active", func() {
		Expect(registry.SetLevel("xds", kuma_log.DebugLevel)).To(Succeed())
		Expect(registry.SetLevel("kds", kuma_log.InfoLevel)).To(Succeed())

		removed := registry.ResetAll()
		Expect(removed).To(HaveLen(2))
		Expect(removed["xds"]).To(Equal(kuma_log.DebugLevel))
		Expect(removed["kds"]).To(Equal(kuma_log.InfoLevel))
	})

	It("ResetAll when empty returns empty map", func() {
		removed := registry.ResetAll()
		Expect(removed).To(BeEmpty())
	})
})

var _ = Describe("ValidateComponentName", func() {
	DescribeTable("valid names",
		func(name string) {
			Expect(kuma_log.ValidateComponentName(name)).To(Succeed())
		},
		Entry("simple", "xds"),
		Entry("dot-separated", "xds.server"),
		Entry("deep hierarchy", "plugins.authn.api-server.tokens"),
		Entry("with dash", "kds-mux"),
		Entry("with underscore", "xds_server"),
		Entry("starts with digit", "1xds"),
		Entry("all digits", "123"),
		Entry("max length (256)", strings.Repeat("a", 256)),
	)

	DescribeTable("invalid names",
		func(name string) {
			Expect(kuma_log.ValidateComponentName(name)).To(HaveOccurred())
		},
		Entry("empty", ""),
		Entry("exceeds 256 chars", strings.Repeat("a", 257)),
		Entry("starts with dot", ".xds"),
		Entry("starts with dash", "-xds"),
		Entry("contains exclamation", "xds!server"),
		Entry("path traversal", "../../etc"),
		Entry("contains space", "xds server"),
		Entry("contains slash", "xds/server"),
	)
})

var _ = Describe("SplitHierarchy", func() {
	DescribeTable("splits component name into ancestors",
		func(name string, expected []string) {
			Expect(kuma_log.SplitHierarchy(name)).To(Equal(expected))
		},
		Entry("single segment", "xds", []string{"xds"}),
		Entry("two segments", "xds.server", []string{"xds.server", "xds"}),
		Entry("three segments", "plugins.authn.tokens", []string{"plugins.authn.tokens", "plugins.authn", "plugins"}),
		Entry("four segments", "a.b.c.d", []string{"a.b.c.d", "a.b.c", "a.b", "a"}),
	)
})
