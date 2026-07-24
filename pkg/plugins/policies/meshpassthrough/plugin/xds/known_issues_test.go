package xds_test

import (
	matcher_config "github.com/cncf/xds/go/xds/type/matcher/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshpassthrough/plugin/xds"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

// These specs pin down semantic edge cases of the Matcher API path that the structural
// invariants in invariants_test.go cannot catch. Pending (P) specs describe behavior that is
// currently BROKEN and act as regression targets — see FINDINGS for the full write-up.

// buildMatcherYAML configures both passthrough listeners and returns the IPv4 listener's
// FilterChainMatcher rendered as YAML for substring/ordering assertions.
func buildMatcherYAML(conf api.Conf) string {
	ipv4 := &envoy_listener.Listener{Name: "outbound:passthrough:ipv4"}
	ipv6 := &envoy_listener.Listener{Name: "outbound:passthrough:ipv6"}
	rs := core_xds.NewResourceSet()
	c := plugin_xds.Configurer{
		APIVersion:    envoy_common.APIV3,
		Conf:          conf,
		IPv6Enabled:   true,
		UseMatcherAPI: true,
	}
	Expect(c.Configure(ipv4, ipv6, rs)).To(Succeed())
	return matcherToYAML(ipv4.GetFilterChainMatcher())
}

func matcherToYAML(m *matcher_config.Matcher) string {
	if m == nil {
		return ""
	}
	y, err := util_proto.ToYAML(m)
	Expect(err).ToNot(HaveOccurred())
	return string(y)
}

var _ = Describe("MeshPassthrough Matcher API edge cases", func() {
	Describe("wildcard TLS SNI", func() {
		wildcardTLS := conf(mtch("Domain", "*.example.com", "tls", p(443)))

		// A wildcard TLS domain must be expressed with a ServerNameMatcher (custom_match) carrying
		// the literal "*.example.com", which Envoy matches as a domain suffix. It must NOT use a
		// prefix_match_map on ServerNameInput: that is a literal character-prefix trie and never
		// matches a real subdomain such as "foo.example.com" (it would silently drop all traffic).
		It("uses a ServerNameMatcher with the wildcard domain, not a prefix_match_map", func() {
			y4 := buildMatcherYAML(wildcardTLS)
			Expect(y4).To(ContainSubstring("envoy.matching.custom_matchers.domain_matcher"))
			Expect(y4).To(ContainSubstring("ServerNameMatcher"))
			Expect(y4).To(ContainSubstring("*.example.com"))
			// The SNI is fed from the raw server name, and no broken prefix map is emitted.
			Expect(y4).To(ContainSubstring("ServerNameInput"))
			Expect(y4).ToNot(ContainSubstring("prefixMatchMap"))
		})
	})

	Describe("TLS with only IP/CIDR matches (no SNI)", func() {
		tlsIPOnly := conf(
			mtch("IP", "10.0.0.1", "tls", p(443)),
			mtch("CIDR", "192.168.0.0/16", "tls", p(443)),
		)

		// FINDING (LOW/RISK): the "tls" branch wraps the IP/CIDR fallback in a Matcher node that
		// has only on_no_match and no matcher_type. It passes protobuf (PGV) validation, but
		// whether Envoy's C++ matcher factory accepts a type-less Matcher node is version
		// dependent and is not covered by any e2e test. This spec locks in the current shape so a
		// change is noticed; Envoy acceptance still needs an e2e check on the target Envoy build.
		It("produces an on_no_match-only node that still passes proto validation", func() {
			y4 := buildMatcherYAML(tlsIPOnly)
			// The tls branch immediately falls through to an IP MatcherList via on_no_match.
			Expect(y4).To(ContainSubstring("onNoMatch"))
			Expect(y4).To(ContainSubstring("envoy.matching.matchers.ip"))
			// The whole tree validates.
			ipv4 := &envoy_listener.Listener{Name: "l4"}
			ipv6 := &envoy_listener.Listener{Name: "l6"}
			rs := core_xds.NewResourceSet()
			c := plugin_xds.Configurer{APIVersion: envoy_common.APIV3, Conf: tlsIPOnly, IPv6Enabled: true, UseMatcherAPI: true}
			Expect(c.Configure(ipv4, ipv6, rs)).To(Succeed())
			if v, ok := any(ipv4.GetFilterChainMatcher()).(interface{ ValidateAll() error }); ok {
				Expect(v.ValidateAll()).To(Succeed())
			}
		})
	})

	Describe("exact + wildcard TLS SNI on the same port", func() {
		// Both the exact and the wildcard domain go into the same ServerNameMatcher. Envoy's
		// domain matcher does longest-suffix matching, so the exact domain wins over the wildcard
		// automatically regardless of the order we emit them in — we do not need to nest an exact
		// map ahead of a wildcard map by hand.
		It("emits both domains in a single ServerNameMatcher", func() {
			y4 := buildMatcherYAML(conf(
				mtch("Domain", "exact.example.com", "tls", p(443)),
				mtch("Domain", "*.example.com", "tls", p(443)),
			))
			Expect(y4).To(ContainSubstring("ServerNameMatcher"))
			Expect(y4).To(ContainSubstring("exact.example.com"))
			Expect(y4).To(ContainSubstring("*.example.com"))
			Expect(y4).ToNot(ContainSubstring("prefixMatchMap"))
		})
	})
})
