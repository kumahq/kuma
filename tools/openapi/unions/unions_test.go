package unions

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Find", func() {
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
		sites := Find(loadBalancer("roundRobin", "urlRewrite"), nil)

		Expect(sites).To(HaveLen(1))
		Expect(sites[0].Branches).To(Equal([]Branch{
			{Value: "RoundRobin", Property: "roundRobin"},
			{Value: "URLRewrite", Property: "urlRewrite"},
		}))
		Expect(sites[0].Shared).To(BeEmpty())
	})

	It("should ignore an enum whose values do not all name a sibling property", func() {
		// A plain status/mode enum must not be mistaken for a union.
		Expect(Find(loadBalancer("roundRobin"), nil)).To(BeEmpty())
	})

	It("should find nested unions and report their path", func() {
		schema := map[string]any{
			"properties": map[string]any{
				"spec": loadBalancer("roundRobin", "urlRewrite"),
			},
		}

		sites := Find(schema, []string{"properties"})

		Expect(sites).To(HaveLen(1))
		Expect(sites[0].Path).To(Equal([]string{"properties", "properties", "spec"}))
	})

	It("should find a union nested in an array item and keep its shared properties", func() {
		// MeshHTTPRoute nests `filters` inside `backendRefs` items.
		filters := func() map[string]any {
			return map[string]any{
				"type": "array",
				"items": map[string]any{"properties": map[string]any{
					"type":              map[string]any{"enum": []any{"URLRewrite", "RequestMirror"}},
					"urlRewrite":        map[string]any{"type": "object"},
					"requestMirror":     map[string]any{"type": "object"},
					"requestHeaderName": map[string]any{"type": "string"},
				}},
			}
		}
		schema := map[string]any{
			"properties": map[string]any{
				"backendRefs": map[string]any{
					"type": "array",
					"items": map[string]any{"properties": map[string]any{
						"name":    map[string]any{"type": "string"},
						"filters": filters(),
					}},
				},
			},
		}

		sites := Find(schema, nil)

		Expect(sites).To(HaveLen(1))
		Expect(sites[0].Path).To(Equal([]string{"properties", "backendRefs", "items", "properties", "filters", "items"}))
		Expect(sites[0].Branches).To(Equal([]Branch{
			{Value: "URLRewrite", Property: "urlRewrite"},
			{Value: "RequestMirror", Property: "requestMirror"},
		}))
		Expect(sites[0].Shared).To(Equal([]string{"requestHeaderName"}))
	})
})

var _ = Describe("Assignments", func() {
	It("should render nothing when the schema has no union", func() {
		expr, err := Assignments(map[string]any{"properties": map[string]any{"name": map[string]any{"type": "string"}}}, nil)

		Expect(err).ToNot(HaveOccurred())
		Expect(expr).To(BeEmpty())
	})

	It("should render a yq assignment rooted at the base path", func() {
		schema := map[string]any{
			"required": []any{"type"},
			"properties": map[string]any{
				"type":  map[string]any{"enum": []any{"Server", "Agent"}},
				"agent": map[string]any{"type": "object"},
				"server": map[string]any{
					"type": "object",
				},
			},
		}

		expr, err := Assignments(schema, []string{"components", "schemas", "VaultConfig"})

		Expect(err).ToNot(HaveOccurred())
		// The branches follow the order of the discriminator's enum, not the order
		// the variant properties happen to be in, and copy the variant schema by
		// path so it keeps its key order.
		Expect(expr).To(Equal(`."components"."schemas"."VaultConfig".oneOf = [` +
			`{"title": "VaultConfigServer", "type": "object", "required": ["type"], "properties": {` +
			`"type": {"type": "string", "enum": ["Server"], "const": "Server"}, ` +
			`"server": ."components"."schemas"."VaultConfig"."properties"."server"}}, ` +
			`{"title": "VaultConfigAgent", "type": "object", "required": ["type"], "properties": {` +
			`"type": {"type": "string", "enum": ["Agent"], "const": "Agent"}, ` +
			`"agent": ."components"."schemas"."VaultConfig"."properties"."agent"}}]`))
	})

	It("should keep the discriminator description and the required shared properties", func() {
		schema := map[string]any{
			"required": []any{"type", "name"},
			"properties": map[string]any{
				"type": map[string]any{"description": "Type of the backend.", "enum": []any{"Tcp", "File"}},
				"name": map[string]any{"type": "string"},
				"tcp":  map[string]any{"type": "object"},
				"file": map[string]any{"type": "object"},
			},
		}

		expr, err := Assignments(schema, []string{"b"})

		Expect(err).ToNot(HaveOccurred())
		Expect(expr).To(ContainSubstring(`{"title": "BTcp", "type": "object", "required": ["type","name"], "properties": {` +
			`"type": {"description": "Type of the backend.", "type": "string", "enum": ["Tcp"], "const": "Tcp"}, ` +
			`"tcp": ."b"."properties"."tcp", "name": ."b"."properties"."name"}}`))
	})

	It("should rewrite nested unions before the union enclosing them", func() {
		// MeshAccessLog nests a `format` union inside the `file` variant, so the
		// enclosing branch has to copy the already rewritten `format`.
		schema := map[string]any{
			"properties": map[string]any{
				"type": map[string]any{"enum": []any{"File", "Tcp"}},
				"tcp":  map[string]any{"type": "object"},
				"file": map[string]any{"properties": map[string]any{
					"format": map[string]any{"properties": map[string]any{
						"type":  map[string]any{"enum": []any{"Plain", "Json"}},
						"plain": map[string]any{"type": "string"},
						"json":  map[string]any{"type": "array"},
					}},
				}},
			},
		}

		expr, err := Assignments(schema, nil)

		Expect(err).ToNot(HaveOccurred())
		inner := strings.Index(expr, `."properties"."file"."properties"."format".oneOf = `)
		outer := strings.Index(expr, "\n  | .oneOf = ")
		Expect(inner).To(BeNumerically(">=", 0))
		Expect(outer).To(BeNumerically(">", inner))
	})
})

var _ = Describe("field", func() {
	DescribeTable("should name the property holding the union",
		func(path []string, expected string) {
			Expect(Site{Path: path}.field()).To(Equal(expected))
		},
		Entry("object property", []string{"properties", "spec", "properties", "default", "properties", "loadBalancer"}, "loadBalancer"),
		Entry("array item", []string{"properties", "backendRefs", "items", "properties", "filters", "items"}, "filters"),
		Entry("schema of its own", []string{"components", "schemas", "VaultConfig"}, "vaultConfig"),
		Entry("root", nil, ""),
	)
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
