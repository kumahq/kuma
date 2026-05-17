package inbound_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/inbound"
)

var _ = Describe("SortRules", func() {
	It("sorts rules by match specificity", func() {
		// given
		rules := []*inbound.Rule{
			catchAllRule("catch-all"),
			sniRule("sni", "backend.mesh"),
			spiffeRule("spiffe-prefix", common_api.PrefixMatchType, "spiffe://default/ns/backend"),
			spiffeRule("spiffe-exact", common_api.ExactMatchType, "spiffe://default/ns/backend/sa/app"),
		}

		// when
		inbound.SortRules(rules)

		// then
		Expect(ruleNames(rules)).To(Equal([]string{
			"spiffe-exact",
			"spiffe-prefix",
			"sni",
			"catch-all",
		}))
	})

	It("keeps stable order for rules with the same match specificity", func() {
		// given
		rules := []*inbound.Rule{
			spiffeRule("first-exact", common_api.ExactMatchType, "spiffe://default/ns/backend/sa/first"),
			spiffeRule("second-exact", common_api.ExactMatchType, "spiffe://default/ns/backend/sa/second"),
			sniRule("first-sni", "first.mesh"),
			sniRule("second-sni", "second.mesh"),
			catchAllRule("first-catch-all"),
			catchAllRule("second-catch-all"),
		}

		// when
		inbound.SortRules(rules)

		// then
		Expect(ruleNames(rules)).To(Equal([]string{
			"first-exact",
			"second-exact",
			"first-sni",
			"second-sni",
			"first-catch-all",
			"second-catch-all",
		}))
	})
})

func spiffeRule(name string, matchType common_api.SpiffeIDMatchType, value string) *inbound.Rule {
	return &inbound.Rule{
		Matches: []common_api.Match{{
			SpiffeID: &common_api.SpiffeIDMatch{
				Type:  matchType,
				Value: value,
			},
		}},
		Conf: name,
	}
}

func sniRule(name string, value string) *inbound.Rule {
	return &inbound.Rule{
		Matches: []common_api.Match{{
			SNI: &common_api.SNIMatch{
				Type:  common_api.SNIExactMatchType,
				Value: value,
			},
		}},
		Conf: name,
	}
}

func catchAllRule(name string) *inbound.Rule {
	return &inbound.Rule{Conf: name}
}

func ruleNames(rules []*inbound.Rule) []string {
	names := make([]string, 0, len(rules))
	for _, rule := range rules {
		names = append(names, rule.Conf.(string))
	}
	return names
}
