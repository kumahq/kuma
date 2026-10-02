package unions

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"
)

// object is a union-shaped schema node: a required `type` enum plus the given
// properties.
func object(values []any, properties map[string]any) map[string]any {
	properties["type"] = map[string]any{"enum": values, "type": "string"}
	return map[string]any{
		"description": "a union",
		"type":        "object",
		"required":    []any{"type"},
		"properties":  properties,
	}
}

func filter() map[string]any {
	return object([]any{"RequestRedirect", "URLRewrite"}, map[string]any{
		"requestRedirect": map[string]any{"type": "object", "properties": map[string]any{
			"path": object([]any{"ReplaceFullPath", "ReplacePrefixMatch"}, map[string]any{
				"replaceFullPath":    map[string]any{"type": "string"},
				"replacePrefixMatch": map[string]any{"type": "string"},
			}),
		}},
		"urlRewrite": map[string]any{"type": "object"},
	})
}

// route has `filters` both on a rule and inside its backendRefs, the way
// MeshHTTPRoute does, and a union inside a variant of the filter union.
func route() map[string]any {
	return map[string]any{
		"spec": map[string]any{"properties": map[string]any{
			"filters": map[string]any{"type": "array", "items": filter()},
			"backendRefs": map[string]any{"type": "array", "items": map[string]any{
				"properties": map[string]any{
					"filters": map[string]any{"type": "array", "items": filter()},
				},
			}},
		}},
	}
}

func backends(variants ...string) map[string]any {
	properties := map[string]any{}
	values := make([]any, 0, len(variants))
	for _, v := range variants {
		properties[lowerFirst(v)] = map[string]any{"type": "object"}
		values = append(values, v)
	}
	return map[string]any{
		"spec": map[string]any{"properties": map[string]any{
			"backends": map[string]any{"type": "array", "items": object(values, properties)},
		}},
	}
}

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
		Expect(sites[0].Values).To(Equal([]string{"RoundRobin", "URLRewrite"}))
		Expect(sites[0].Variants).To(Equal(map[string]string{"RoundRobin": "roundRobin", "URLRewrite": "urlRewrite"}))
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
})

var _ = Describe("Plan", func() {
	It("should give every discriminator value a mapping entry and a member schema", func() {
		schema := map[string]any{"loadBalancer": object([]any{"RoundRobin", "RingHash"}, map[string]any{
			"roundRobin": map[string]any{"type": "object"},
			"ringHash":   map[string]any{"type": "object", "properties": map[string]any{"minRingSize": map[string]any{"type": "integer"}}},
		})}

		rewrite, err := Plan(schema, []string{"properties"}, "MeshLoadBalancingStrategy")

		Expect(err).ToNot(HaveOccurred())
		Expect(rewrite.Unions).To(HaveLen(1))
		u := rewrite.Unions["properties.loadBalancer"]
		Expect(u.OneOf).To(Equal([]any{
			map[string]any{"$ref": "#/components/schemas/MeshLoadBalancingStrategyLoadBalancerRoundRobin"},
			map[string]any{"$ref": "#/components/schemas/MeshLoadBalancingStrategyLoadBalancerRingHash"},
		}))
		Expect(u.Discriminator).To(Equal(map[string]any{
			"propertyName": "type",
			"mapping": map[string]any{
				"RoundRobin": "#/components/schemas/MeshLoadBalancingStrategyLoadBalancerRoundRobin",
				"RingHash":   "#/components/schemas/MeshLoadBalancingStrategyLoadBalancerRingHash",
			},
		}))
		Expect(rewrite.Members).To(HaveLen(2))
		Expect(rewrite.Members["MeshLoadBalancingStrategyLoadBalancerRingHash"]).To(Equal(Member{
			Title:    "RingHash",
			Type:     "object",
			Required: []any{"type"},
			Properties: map[string]any{
				"type":     map[string]any{"enum": []any{"RingHash"}, "type": "string"},
				"ringHash": map[string]any{"type": "object", "properties": map[string]any{"minRingSize": map[string]any{"type": "integer"}}},
			},
		}))
	})

	// `type: RoundRobin` on its own is a complete configuration.
	It("should not require the variant property, so a variant with no config stays valid", func() {
		schema := map[string]any{"loadBalancer": object([]any{"RoundRobin", "Random"}, map[string]any{
			"roundRobin": map[string]any{"type": "object"},
			"random":     map[string]any{"type": "object"},
		})}

		rewrite, err := Plan(schema, nil, "P")

		Expect(err).ToNot(HaveOccurred())
		for _, member := range rewrite.Members {
			Expect(member.Required).To(Equal([]any{"type"}))
		}
	})

	It("should keep the properties the variants share in every member", func() {
		schema := map[string]any{"hashPolicies": map[string]any{"type": "array", "items": object([]any{"Header", "Cookie"}, map[string]any{
			"header":   map[string]any{"type": "object"},
			"cookie":   map[string]any{"type": "object"},
			"terminal": map[string]any{"type": "boolean"},
		})}}

		rewrite, err := Plan(schema, []string{"properties"}, "P")

		Expect(err).ToNot(HaveOccurred())
		Expect(rewrite.Members).To(HaveKey("PHashPolicyHeader"))
		Expect(rewrite.Members["PHashPolicyHeader"].Properties).To(HaveKey("terminal"))
		Expect(rewrite.Members["PHashPolicyHeader"].Properties).ToNot(HaveKey("cookie"))
		Expect(rewrite.Members["PHashPolicyCookie"].Properties).To(HaveKey("terminal"))
	})

	It("should resolve an acronym value to its property and keep it in the member name", func() {
		rewrite, err := Plan(map[string]any{"filter": filter()}, []string{"properties"}, "MeshHTTPRoute")

		Expect(err).ToNot(HaveOccurred())
		Expect(rewrite.Members).To(HaveKey("MeshHTTPRouteFilterURLRewrite"))
		Expect(rewrite.Members["MeshHTTPRouteFilterURLRewrite"].Properties).To(HaveKey("urlRewrite"))
		Expect(rewrite.Members["MeshHTTPRouteFilterURLRewrite"].Properties).ToNot(HaveKey("requestRedirect"))
	})

	It("should describe nested unions inside the members of the outer union", func() {
		rewrite, err := Plan(route(), []string{"properties"}, "MeshHTTPRoute")

		Expect(err).ToNot(HaveOccurred())
		// Both `filters` are the same type, so they share their members.
		Expect(rewrite.Unions).To(HaveLen(2))
		Expect(rewrite.Unions).To(HaveKey("properties.spec.properties.filters.items"))
		Expect(rewrite.Unions).To(HaveKey("properties.spec.properties.backendRefs.items.properties.filters.items"))
		Expect(rewrite.Unions["properties.spec.properties.filters.items"].OneOf).To(
			Equal(rewrite.Unions["properties.spec.properties.backendRefs.items.properties.filters.items"].OneOf))
		Expect(rewrite.Members).To(HaveLen(4))

		// The path union inside requestRedirect is not a union of the document any
		// more, the member holding requestRedirect refers to its members.
		redirect := rewrite.Members["MeshHTTPRouteFilterRequestRedirect"].Properties["requestRedirect"].(map[string]any)
		path := redirect["properties"].(map[string]any)["path"].(map[string]any)
		Expect(path).ToNot(HaveKey("properties"))
		Expect(path["discriminator"]).To(HaveKeyWithValue("mapping", map[string]any{
			"ReplaceFullPath":    "#/components/schemas/MeshHTTPRoutePathReplaceFullPath",
			"ReplacePrefixMatch": "#/components/schemas/MeshHTTPRoutePathReplacePrefixMatch",
		}))
		Expect(rewrite.Members["MeshHTTPRoutePathReplaceFullPath"].Properties).To(HaveKey("replaceFullPath"))
	})

	It("should qualify the names of two different unions under the same property", func() {
		schema := map[string]any{
			"rules": map[string]any{"properties": backends("Tcp", "File")},
			"to":    map[string]any{"properties": backends("Tcp", "OpenTelemetry")},
		}
		// Equal members would be shared, so make one of the clashing names differ.
		dig(schema, []string{"to", "properties", "spec", "properties", "backends", "items", "properties"}).(map[string]any)["tcp"] = map[string]any{"type": "string"}

		rewrite, err := Plan(schema, []string{"properties"}, "P")

		Expect(err).ToNot(HaveOccurred())
		Expect(rewrite.Members).To(HaveKey("PRulesSpecBackendTcp"))
		Expect(rewrite.Members).To(HaveKey("PToSpecBackendTcp"))
		Expect(rewrite.Members).To(HaveKey("PRulesSpecBackendFile"))
		Expect(rewrite.Members).To(HaveKey("PToSpecBackendOpenTelemetry"))
	})

	It("should keep the names of two policies with a backends union apart", func() {
		trace, err := Plan(backends("Zipkin", "OpenTelemetry"), []string{"properties"}, "MeshTrace")
		Expect(err).ToNot(HaveOccurred())
		accessLog, err := Plan(backends("File", "OpenTelemetry"), []string{"properties"}, "MeshAccessLog")
		Expect(err).ToNot(HaveOccurred())

		Expect(trace.Members).To(HaveKey("MeshTraceBackendOpenTelemetry"))
		Expect(accessLog.Members).To(HaveKey("MeshAccessLogBackendOpenTelemetry"))
		for name := range trace.Members {
			Expect(accessLog.Members).ToNot(HaveKey(name))
		}
	})

	It("should describe the loadBalancer union of the generated MeshLoadBalancingStrategy spec", func() {
		raw, err := os.ReadFile("../../../pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1/rest.yaml")
		Expect(err).ToNot(HaveOccurred())
		var spec map[string]any
		Expect(yaml.Unmarshal(raw, &spec)).To(Succeed())
		item := dig(spec, []string{"components", "schemas", "MeshLoadBalancingStrategyItem"})

		rewrite, err := Plan(item, []string{"components", "schemas", "MeshLoadBalancingStrategyItem"}, "MeshLoadBalancingStrategy")

		Expect(err).ToNot(HaveOccurred())
		loadBalancer := rewrite.Unions["components.schemas.MeshLoadBalancingStrategyItem.properties.spec.properties.to.items.properties.default.properties.loadBalancer"]
		Expect(loadBalancer.Discriminator).To(HaveKeyWithValue("mapping", map[string]any{
			"RoundRobin":   "#/components/schemas/MeshLoadBalancingStrategyLoadBalancerRoundRobin",
			"LeastRequest": "#/components/schemas/MeshLoadBalancingStrategyLoadBalancerLeastRequest",
			"RingHash":     "#/components/schemas/MeshLoadBalancingStrategyLoadBalancerRingHash",
			"Random":       "#/components/schemas/MeshLoadBalancingStrategyLoadBalancerRandom",
			"Maglev":       "#/components/schemas/MeshLoadBalancingStrategyLoadBalancerMaglev",
		}))
		Expect(rewrite.Members["MeshLoadBalancingStrategyLoadBalancerRingHash"].Properties).To(HaveKey("ringHash"))
	})
})

var _ = Describe("Assignments", func() {
	It("should render nothing when the schema has no union", func() {
		expr, err := Assignments(map[string]any{"properties": map[string]any{"name": map[string]any{"type": "string"}}}, nil, "P")

		Expect(err).ToNot(HaveOccurred())
		Expect(expr).To(BeEmpty())
	})

	It("should add the members and point the union rooted at the base path at them", func() {
		schema := map[string]any{
			"properties": map[string]any{
				"type":  map[string]any{"enum": []any{"Server", "Agent"}},
				"agent": map[string]any{"type": "object"},
				"server": map[string]any{
					"type": "object",
				},
			},
		}

		expr, err := Assignments(schema, []string{"components", "schemas", "VaultConfig"}, "VaultConfig")

		Expect(err).ToNot(HaveOccurred())
		// The members are sorted by name, the oneOf follows the order of the
		// discriminator's enum.
		Expect(expr).To(Equal(`."components"."schemas"."VaultConfigAgent" = ` +
			`{"title":"Agent","type":"object","required":["type"],"properties":{"agent":{"type":"object"},"type":{"enum":["Agent"]}}}` + "\n  | " +
			`."components"."schemas"."VaultConfigServer" = ` +
			`{"title":"Server","type":"object","required":["type"],"properties":{"server":{"type":"object"},"type":{"enum":["Server"]}}}` + "\n  | " +
			`."components"."schemas"."VaultConfig" |= (del(.properties, .required, .type)` +
			` | .oneOf = [{"$ref":"#/components/schemas/VaultConfigServer"},{"$ref":"#/components/schemas/VaultConfigAgent"}]` +
			` | .discriminator = {"mapping":{"Agent":"#/components/schemas/VaultConfigAgent","Server":"#/components/schemas/VaultConfigServer"},"propertyName":"type"})`))
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

var _ = Describe("Patch", func() {
	write := func(content string) string {
		path := filepath.Join(GinkgoT().TempDir(), "rest.yaml")
		Expect(os.WriteFile(path, []byte(content), 0o600)).To(Succeed())
		return path
	}

	// An already patched document has no union left to find, so running it twice
	// must not even start yq.
	It("should leave a document with no union alone", func() {
		path := write("components:\n  schemas:\n    FooItem:\n      properties:\n        name:\n          type: string\n")

		Expect(Patch(context.Background(), path, "/nonexistent/yq", GinkgoWriter)).To(Succeed())
	})

	It("should refuse to replace an existing schema with a member", func() {
		path := write(`components:
  schemas:
    FooItem:
      properties:
        mode:
          properties:
            type: {enum: [A, B]}
            a: {type: object}
            b: {type: object}
    FooModeA:
      type: string
`)

		err := Patch(context.Background(), path, "/nonexistent/yq", GinkgoWriter)

		Expect(err).To(MatchError(ContainSubstring("union member FooModeA would replace a schema of the same name")))
	})
})
