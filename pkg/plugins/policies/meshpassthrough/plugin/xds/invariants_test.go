package xds_test

import (
	"fmt"

	matcher_config "github.com/cncf/xds/go/xds/type/matcher/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshpassthrough/plugin/xds"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

// The invariants below are the structural guarantees a MeshPassthrough listener must satisfy
// for Envoy to accept it and route deterministically. They are asserted for EVERY generated
// config, across both the legacy (per-filter-chain FilterChainMatch) and the Matcher API
// (Listener.FilterChainMatcher) code paths.
//
// Why these matter:
//   - Duplicate filter chain names  -> Envoy rejects the whole listener.
//   - A matcher tree action pointing at a non-existent chain (dangling ref) -> the connection
//     hits `no_filter_chain_match` and is silently dropped.
//   - A filter chain that no matcher path ever selects (orphan) -> dead config; usually a symptom
//     of a naming/expansion bug.
//   - The two code paths MUST produce the same set of chains; only the selection mechanism differs.

const sentinelChain = "SENTINEL_UNTOUCHED"

func newBareListener(name string) *envoy_listener.Listener {
	return &envoy_listener.Listener{
		Name:         name,
		FilterChains: []*envoy_listener.FilterChain{{Name: sentinelChain}},
	}
}

func wasConfigured(l *envoy_listener.Listener) bool {
	return len(l.FilterChains) != 1 || l.FilterChains[0].Name != sentinelChain
}

func chainNameSet(l *envoy_listener.Listener) map[string]int {
	out := map[string]int{}
	for _, fc := range l.FilterChains {
		out[fc.GetName()]++
	}
	return out
}

// collectReferencedChains walks a FilterChainMatcher tree and returns, for each filter-chain
// name, how many distinct actions reference it.
func collectReferencedChains(m *matcher_config.Matcher) map[string]int {
	acc := map[string]int{}
	walkMatcher(m, acc)
	return acc
}

func walkMatcher(m *matcher_config.Matcher, acc map[string]int) {
	if m == nil {
		return
	}
	switch mt := m.MatcherType.(type) {
	case *matcher_config.Matcher_MatcherList_:
		for _, fm := range mt.MatcherList.GetMatchers() {
			walkOnMatch(fm.GetOnMatch(), acc)
		}
	case *matcher_config.Matcher_MatcherTree_:
		tree := mt.MatcherTree
		if em := tree.GetExactMatchMap(); em != nil {
			for _, om := range em.GetMap() {
				walkOnMatch(om, acc)
			}
		}
		if pm := tree.GetPrefixMatchMap(); pm != nil {
			for _, om := range pm.GetMap() {
				walkOnMatch(om, acc)
			}
		}
		if cm := tree.GetCustomMatch(); cm != nil {
			// The domain (SNI) custom matcher hides its on-matches inside a marshaled
			// ServerNameMatcher; descend into it so its chains are counted as referenced.
			snm := &matcher_config.ServerNameMatcher{}
			if err := cm.GetTypedConfig().UnmarshalTo(snm); err == nil {
				for _, dm := range snm.GetDomainMatchers() {
					walkOnMatch(dm.GetOnMatch(), acc)
				}
			}
		}
	}
	walkOnMatch(m.GetOnNoMatch(), acc)
}

func walkOnMatch(om *matcher_config.Matcher_OnMatch, acc map[string]int) {
	if om == nil {
		return
	}
	switch t := om.OnMatch.(type) {
	case *matcher_config.Matcher_OnMatch_Matcher:
		walkMatcher(t.Matcher, acc)
	case *matcher_config.Matcher_OnMatch_Action:
		acc[t.Action.GetName()]++
	}
}

func keys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// configure runs the Configurer over a fresh pair of listeners.
func configure(conf api.Conf, useMatcherAPI bool) (*envoy_listener.Listener, *envoy_listener.Listener, error) {
	ipv4 := newBareListener("outbound:passthrough:ipv4")
	ipv6 := newBareListener("outbound:passthrough:ipv6")
	rs := core_xds.NewResourceSet()
	c := plugin_xds.Configurer{
		APIVersion:    envoy_common.APIV3,
		Conf:          conf,
		IPv6Enabled:   true,
		UseMatcherAPI: useMatcherAPI,
	}
	err := c.Configure(ipv4, ipv6, rs)
	return ipv4, ipv6, err
}

func assertListenerInvariants(desc string, l *envoy_listener.Listener, useMatcherAPI bool) {
	names := chainNameSet(l)

	// (1) Filter chain names are unique. Envoy rejects a listener with duplicates.
	for name, count := range names {
		Expect(count).To(Equal(1), "%s: duplicate filter chain name %q (x%d)", desc, name, count)
	}

	if useMatcherAPI {
		// (2) With the Matcher API a FilterChainMatcher must exist whenever there are chains,
		//     and individual chains must NOT carry a FilterChainMatch (selection is centralized).
		Expect(l.GetFilterChainMatcher()).ToNot(BeNil(), "%s: chains present but no FilterChainMatcher", desc)
		for _, fc := range l.FilterChains {
			Expect(fc.GetFilterChainMatch()).To(BeNil(),
				"%s: chain %q has a FilterChainMatch under the Matcher API", desc, fc.GetName())
		}

		// (2b) The generated matcher tree must pass protobuf (PGV) validation, otherwise
		//      go-control-plane / the xDS server rejects the snapshot before it reaches Envoy.
		if v, ok := any(l.GetFilterChainMatcher()).(interface{ ValidateAll() error }); ok {
			Expect(v.ValidateAll()).To(Succeed(), "%s: FilterChainMatcher failed proto validation", desc)
		}

		referenced := collectReferencedChains(l.GetFilterChainMatcher())

		// (3) No dangling references: every chain the matcher tree can select must exist.
		for ref := range referenced {
			Expect(names).To(HaveKey(ref),
				"%s: matcher tree references chain %q which does not exist (chains=%v)", desc, ref, keys(names))
		}
		// (4) No orphans: every chain must be selectable by the matcher tree.
		for name := range names {
			Expect(referenced).To(HaveKey(name),
				"%s: chain %q is never referenced by the matcher tree (referenced=%v)", desc, name, keys(referenced))
		}
	} else {
		// Legacy path: no centralized matcher, every chain selects itself via a FilterChainMatch.
		Expect(l.GetFilterChainMatcher()).To(BeNil(), "%s: legacy path unexpectedly set a FilterChainMatcher", desc)
		for _, fc := range l.FilterChains {
			Expect(fc.GetFilterChainMatch()).ToNot(BeNil(),
				"%s: legacy chain %q has no FilterChainMatch", desc, fc.GetName())
		}
	}
}

func invariantBody() func(conf api.Conf) {
	return func(conf api.Conf) {
		// Legacy path.
		l4Legacy, l6Legacy, errLegacy := configure(conf, false)
		Expect(errLegacy).ToNot(HaveOccurred(), "legacy Configure returned an error")

		// Matcher API path.
		l4Matcher, l6Matcher, errMatcher := configure(conf, true)
		Expect(errMatcher).ToNot(HaveOccurred(), "matcher-API Configure returned an error")

		for _, tc := range []struct {
			desc    string
			legacy  *envoy_listener.Listener
			matcher *envoy_listener.Listener
		}{
			{"ipv4", l4Legacy, l4Matcher},
			{"ipv6", l6Legacy, l6Matcher},
		} {
			legacyConfigured := wasConfigured(tc.legacy)
			matcherConfigured := wasConfigured(tc.matcher)

			// (5) Both paths must decide identically whether this listener is in play.
			Expect(matcherConfigured).To(Equal(legacyConfigured),
				"%s: legacy configured=%v but matcher configured=%v", tc.desc, legacyConfigured, matcherConfigured)

			if !legacyConfigured {
				continue
			}

			assertListenerInvariants(tc.desc, tc.legacy, false)
			assertListenerInvariants(tc.desc, tc.matcher, true)

			// (6) The two paths must emit the SAME set of filter chains; only the
			//     selection mechanism (per-chain match vs central matcher) differs.
			Expect(keys(chainNameSet(tc.matcher))).To(ConsistOf(keys(chainNameSet(tc.legacy))),
				"%s: matcher-API chains differ from legacy chains", tc.desc)
		}
	}
}

var _ = Describe("MeshPassthrough listener invariants", func() {
	args := append([]any{invariantBody()}, invariantEntries()...)
	DescribeTable("hold for every generated config on both code paths", args...)
})

// --- config matrix -----------------------------------------------------------------------------

func mtch(t, v, proto string, port *uint32) api.Match {
	return api.Match{
		Type:     api.MatchType(t),
		Value:    v,
		Protocol: api.ProtocolType(proto),
		Port:     port,
	}
}

func p(v uint32) *uint32 { return &v }

func conf(matches ...api.Match) api.Conf {
	m := matches
	return api.Conf{
		PassthroughMode: matchedMode(),
		AppendMatch:     &m,
	}
}

func matchedMode() *api.PassthroughMode {
	pm := api.PassthroughMode("Matched")
	return &pm
}

func invariantEntries() []any {
	var entries []any
	add := func(name string, c api.Conf) {
		entries = append(entries, Entry(name, c))
	}

	// --- A. single-match coverage: each protocol x value-kind x port variant ------------------
	l7 := []string{"http", "http2", "grpc"}
	l4 := []string{"tcp", "tls"}

	for _, port := range []*uint32{p(8080), nil} {
		ps := "all-ports"
		if port != nil {
			ps = fmt.Sprintf("port-%d", *port)
		}
		// Domains: L7 + TLS only (tcp/mysql domains are rejected by validation).
		for _, proto := range append(append([]string{}, l7...), "tls") {
			add(fmt.Sprintf("single/domain/%s/%s", proto, ps),
				conf(mtch("Domain", "example.com", proto, port)))
			add(fmt.Sprintf("single/wildcard-domain/%s/%s", proto, ps),
				conf(mtch("Domain", "*.example.com", proto, port)))
		}
		// IPs and CIDRs: all protocols except domain-only ones. tcp/tls/L7 all valid.
		for _, proto := range append(append([]string{}, l4...), l7...) {
			add(fmt.Sprintf("single/ipv4/%s/%s", proto, ps),
				conf(mtch("IP", "10.0.0.1", proto, port)))
			add(fmt.Sprintf("single/ipv6/%s/%s", proto, ps),
				conf(mtch("IP", "fd00::1", proto, port)))
			add(fmt.Sprintf("single/cidrv4/%s/%s", proto, ps),
				conf(mtch("CIDR", "10.0.0.0/8", proto, port)))
			add(fmt.Sprintf("single/cidrv6/%s/%s", proto, ps),
				conf(mtch("CIDR", "fd00::/16", proto, port)))
		}
	}
	// MySQL requires a port.
	for _, vk := range []struct{ t, v string }{{"IP", "10.0.0.1"}, {"IP", "fd00::1"}, {"CIDR", "10.0.0.0/8"}, {"CIDR", "fd00::/16"}} {
		add(fmt.Sprintf("single/mysql/%s", vk.v), conf(mtch(vk.t, vk.v, "mysql", p(3306))))
	}

	// --- B. interaction cases -----------------------------------------------------------------

	// Dual-stack same protocol.
	add("dualstack/ip/tcp", conf(
		mtch("IP", "10.0.0.1", "tcp", p(80)),
		mtch("IP", "fd00::1", "tcp", p(80)),
	))
	add("dualstack/cidr/tls", conf(
		mtch("CIDR", "10.0.0.0/8", "tls", p(443)),
		mtch("CIDR", "fd00::/16", "tls", p(443)),
	))

	// Same IP, multiple ports (must collapse into one predicate with a port sub-tree).
	add("same-ip/multi-port/tcp", conf(
		mtch("IP", "10.0.0.1", "tcp", p(80)),
		mtch("IP", "10.0.0.1", "tcp", p(443)),
		mtch("IP", "10.0.0.1", "tcp", p(8080)),
	))

	// Same IP, different protocols on different ports.
	add("same-ip/multi-proto", conf(
		mtch("IP", "10.0.0.1", "tcp", p(80)),
		mtch("IP", "10.0.0.1", "tls", p(443)),
		mtch("IP", "10.0.0.1", "http", p(8080)),
	))

	// Overlapping CIDRs: specific + general (ordering-sensitive).
	add("overlapping-cidr/tcp", conf(
		mtch("CIDR", "10.1.2.0/24", "tcp", p(80)),
		mtch("CIDR", "10.1.0.0/16", "tcp", p(80)),
		mtch("CIDR", "10.0.0.0/8", "tcp", p(80)),
	))

	// IP + CIDR + Domain mixed, HTTP, same port.
	add("mixed/http/same-port", conf(
		mtch("Domain", "a.example.com", "http", p(80)),
		mtch("Domain", "b.example.com", "http", p(80)),
		mtch("IP", "10.0.0.1", "http", p(80)),
		mtch("CIDR", "192.168.0.0/16", "http", p(80)),
	))

	// Port-0 wildcard combined with specific ports (exercises the additional-ports expansion).
	add("wildcard-port/tcp", conf(
		mtch("IP", "10.0.0.1", "tcp", nil),
		mtch("IP", "10.0.0.2", "tcp", p(80)),
		mtch("IP", "10.0.0.3", "tcp", p(443)),
	))
	add("wildcard-port/http-domain", conf(
		mtch("Domain", "wild.example.com", "http", nil),
		mtch("Domain", "specific.example.com", "http", p(8080)),
	))

	// TLS: exact SNI + wildcard SNI + IP fallback.
	add("tls/sni-and-ip", conf(
		mtch("Domain", "exact.example.com", "tls", p(443)),
		mtch("Domain", "*.example.com", "tls", p(443)),
		mtch("IP", "10.0.0.1", "tls", p(443)),
		mtch("CIDR", "192.168.0.0/16", "tls", p(443)),
	))
	// TLS IP-only (no SNI): exercises the SNI-tree with only an IP/CIDR fallback.
	add("tls/ip-only", conf(
		mtch("IP", "10.0.0.1", "tls", p(443)),
		mtch("CIDR", "192.168.0.0/16", "tls", p(443)),
	))
	// TLS wildcard-only.
	add("tls/wildcard-only", conf(
		mtch("Domain", "*.example.com", "tls", p(443)),
	))

	// HTTP many domains + IP on same port (raw_buffer branch: port matcher + IP field matcher).
	add("http/many-domains-and-ip", conf(
		mtch("Domain", "one.example.com", "http", p(80)),
		mtch("Domain", "two.example.com", "http", p(80)),
		mtch("Domain", "three.example.com", "http", p(80)),
		mtch("IP", "10.0.0.1", "http", p(80)),
	))

	// MySQL + TCP on different ports (MySQL chains are named with the tcp protocol).
	add("mysql-and-tcp", conf(
		mtch("IP", "10.0.0.1", "mysql", p(3306)),
		mtch("IP", "10.0.0.1", "tcp", p(80)),
	))

	// Everything at once, dual-stack.
	add("kitchen-sink", conf(
		mtch("Domain", "api.example.com", "tls", p(443)),
		mtch("Domain", "*.example.com", "tls", p(443)),
		mtch("Domain", "web.example.com", "http", p(80)),
		mtch("Domain", "*.svc.local", "http", p(80)),
		mtch("IP", "10.0.0.1", "tcp", p(9000)),
		mtch("IP", "fd00::1", "tcp", p(9000)),
		mtch("CIDR", "192.168.0.0/16", "http", p(80)),
		mtch("CIDR", "fd00::/16", "tls", p(443)),
		mtch("IP", "10.0.0.2", "mysql", p(3306)),
		mtch("Domain", "grpc.example.com", "grpc", p(50051)),
	))

	return entries
}
