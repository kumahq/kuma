package cmd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("findUnionSites", func() {
	loadBalancer := func(variants ...string) map[string]any {
		properties := map[string]any{
			"type": map[string]any{"enum": []any{"RoundRobin", "URLRewrite"}},
		}
		for _, v := range variants {
			properties[v] = map[string]any{"type": "object"}
		}
		return map[string]any{"properties": properties}
	}

	It("should pair every discriminator value with the property it selects", func() {
		sites := findUnionSites(loadBalancer("roundRobin", "urlRewrite"), nil)

		Expect(sites).To(HaveLen(1))
		Expect(sites[0].oneOf).To(Equal([]any{
			map[string]any{"properties": map[string]any{
				"type":       map[string]any{"enum": []any{"RoundRobin"}},
				"roundRobin": map[string]any{},
			}},
			map[string]any{"properties": map[string]any{
				"type":       map[string]any{"enum": []any{"URLRewrite"}},
				"urlRewrite": map[string]any{},
			}},
		}))
	})

	It("should ignore an enum whose values do not all name a sibling property", func() {
		// A plain status/mode enum must not be mistaken for a union.
		Expect(findUnionSites(loadBalancer("roundRobin"), nil)).To(BeEmpty())
	})

	It("should find nested unions and report their path", func() {
		schema := map[string]any{
			"properties": map[string]any{
				"spec": loadBalancer("roundRobin", "urlRewrite"),
			},
		}

		sites := findUnionSites(schema, []string{"properties"})

		Expect(sites).To(HaveLen(1))
		Expect(sites[0].path).To(Equal([]string{"properties", "properties", "spec"}))
	})
})

var _ = Describe("variantProperty", func() {
	DescribeTable("should resolve a discriminator value to its property",
		func(value string, expected string) {
			properties := map[string]any{expected: map[string]any{}}

			actual, ok := variantProperty(properties, value)

			Expect(ok).To(BeTrue())
			Expect(actual).To(Equal(expected))
		},
		Entry("simple", "RoundRobin", "roundRobin"),
		Entry("single leading capital", "Tcp", "tcp"),
		Entry("acronym prefix", "URLRewrite", "urlRewrite"),
		Entry("mixed caps", "OpenTelemetry", "openTelemetry"),
	)

	It("should not resolve to the discriminator itself", func() {
		_, ok := variantProperty(map[string]any{"type": map[string]any{}}, "Type")
		Expect(ok).To(BeFalse())
	})
})
