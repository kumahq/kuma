package xds_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	. "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/plugin/xds"
	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestCEL(t *testing.T) {
	test.RunSpecs(t, "MeshAccessLog CEL")
}

var _ = Describe("MatchToCEL", func() {
	DescribeTable("converts a Match to a CEL expression",
		func(match *common_api.Match, expected string) {
			Expect(MatchToCEL(match)).To(Equal(expected))
		},
		Entry("nil match", (*common_api.Match)(nil), ""),
		Entry("empty match", &common_api.Match{}, ""),
		Entry("spiffeID Exact", &common_api.Match{
			SpiffeID: &common_api.SpiffeIDMatch{
				Type:  common_api.ExactMatchType,
				Value: "spiffe://default/ns/backend-ns/sa/backend",
			},
		}, `connection.uri_san_peer_certificate == "spiffe://default/ns/backend-ns/sa/backend"`),
		Entry("spiffeID Prefix", &common_api.Match{
			SpiffeID: &common_api.SpiffeIDMatch{
				Type:  common_api.PrefixMatchType,
				Value: "spiffe://default/ns/backend-ns/",
			},
		}, `connection.uri_san_peer_certificate.startsWith("spiffe://default/ns/backend-ns/")`),
		Entry("SNI Exact", &common_api.Match{
			SNI: &common_api.SNIMatch{
				Type:  common_api.SNIExactMatchType,
				Value: "sni.extsvc.default.zone-1.aws-aurora.8443",
			},
		}, `connection.requested_server_name == "sni.extsvc.default.zone-1.aws-aurora.8443"`),
		Entry("spiffeID + SNI combined (AND)", &common_api.Match{
			SpiffeID: &common_api.SpiffeIDMatch{
				Type:  common_api.ExactMatchType,
				Value: "spiffe://default/ns/ns-1/sa/sa-1",
			},
			SNI: &common_api.SNIMatch{
				Type:  common_api.SNIExactMatchType,
				Value: "sni.example",
			},
		}, `connection.uri_san_peer_certificate == "spiffe://default/ns/ns-1/sa/sa-1" && connection.requested_server_name == "sni.example"`),
		Entry("value with embedded quote is escaped", &common_api.Match{
			SpiffeID: &common_api.SpiffeIDMatch{
				Type:  common_api.ExactMatchType,
				Value: `spiffe://"weird"/sa`,
			},
		}, `connection.uri_san_peer_certificate == "spiffe://\"weird\"/sa"`),
	)
})

var _ = Describe("ComposeWinnerExpr", func() {
	exactSpiffe := func(v string) *common_api.Match {
		return &common_api.Match{
			SpiffeID: &common_api.SpiffeIDMatch{Type: common_api.ExactMatchType, Value: v},
		}
	}
	prefixSpiffe := func(v string) *common_api.Match {
		return &common_api.Match{
			SpiffeID: &common_api.SpiffeIDMatch{Type: common_api.PrefixMatchType, Value: v},
		}
	}
	sni := func(v string) *common_api.Match {
		return &common_api.Match{
			SNI: &common_api.SNIMatch{Type: common_api.SNIExactMatchType, Value: v},
		}
	}
	exactSpiffeAndSNI := func(spiffeV, sniV string) *common_api.Match {
		return &common_api.Match{
			SpiffeID: &common_api.SpiffeIDMatch{Type: common_api.ExactMatchType, Value: spiffeV},
			SNI:      &common_api.SNIMatch{Type: common_api.SNIExactMatchType, Value: sniV},
		}
	}

	DescribeTable("composes first-match-wins CEL expression",
		func(self *common_api.Match, priors []*common_api.Match, expected string) {
			Expect(ComposeExpr(self, priors)).To(Equal(expected))
		},
		Entry("no self, no priors", (*common_api.Match)(nil), nil, ""),
		Entry("self only, no priors",
			exactSpiffe("spiffe://m/sa1"), nil,
			`connection.uri_san_peer_certificate == "spiffe://m/sa1"`),
		Entry("catch-all with one prior",
			(*common_api.Match)(nil), []*common_api.Match{exactSpiffe("spiffe://m/sa1")},
			`!(connection.uri_san_peer_certificate == "spiffe://m/sa1")`),
		Entry("self with one prior on a different dimension is kept",
			sni("sni-2"), []*common_api.Match{exactSpiffe("spiffe://m/sa1")},
			`connection.requested_server_name == "sni-2" && !(connection.uri_san_peer_certificate == "spiffe://m/sa1")`),
		Entry("priors disjoint on same dimension are dropped; others kept",
			sni("sni-3"),
			[]*common_api.Match{exactSpiffe("spiffe://m/sa1"), sni("sni-2")},
			`connection.requested_server_name == "sni-3" && !(connection.uri_san_peer_certificate == "spiffe://m/sa1")`),
		Entry("nil prior is skipped defensively",
			sni("sni-x"),
			[]*common_api.Match{nil},
			`connection.requested_server_name == "sni-x"`),

		// Disjointness simplification:
		Entry("exact-vs-exact spiffe disjoint prior is dropped",
			exactSpiffe("spiffe://m/sa1"), []*common_api.Match{exactSpiffe("spiffe://m/sa2")},
			`connection.uri_san_peer_certificate == "spiffe://m/sa1"`),
		Entry("exact-vs-exact spiffe equal prior is kept",
			exactSpiffe("spiffe://m/sa1"), []*common_api.Match{exactSpiffe("spiffe://m/sa1")},
			`connection.uri_san_peer_certificate == "spiffe://m/sa1" && !(connection.uri_san_peer_certificate == "spiffe://m/sa1")`),
		Entry("exact self vs non-matching prefix prior is dropped",
			exactSpiffe("spiffe://m/sa1"), []*common_api.Match{prefixSpiffe("spiffe://other/")},
			`connection.uri_san_peer_certificate == "spiffe://m/sa1"`),
		Entry("exact self vs covering prefix prior is kept",
			exactSpiffe("spiffe://m/sa1"), []*common_api.Match{prefixSpiffe("spiffe://m/")},
			`connection.uri_san_peer_certificate == "spiffe://m/sa1" && !(connection.uri_san_peer_certificate.startsWith("spiffe://m/"))`),
		Entry("prefix self vs unrelated prefix prior is dropped",
			prefixSpiffe("spiffe://a/"), []*common_api.Match{prefixSpiffe("spiffe://b/")},
			`connection.uri_san_peer_certificate.startsWith("spiffe://a/")`),
		Entry("prefix self vs comparable prefix prior is kept",
			prefixSpiffe("spiffe://m/ns/"), []*common_api.Match{prefixSpiffe("spiffe://m/")},
			`connection.uri_san_peer_certificate.startsWith("spiffe://m/ns/") && !(connection.uri_san_peer_certificate.startsWith("spiffe://m/"))`),
		Entry("disjoint SNI proves overall disjointness regardless of matching spiffe",
			exactSpiffeAndSNI("spiffe://m/sa1", "sni-1"),
			[]*common_api.Match{exactSpiffeAndSNI("spiffe://m/sa1", "sni-2")},
			`connection.uri_san_peer_certificate == "spiffe://m/sa1" && connection.requested_server_name == "sni-1"`),
		Entry("priors filtered out of a chain leave a clean expression",
			exactSpiffe("spiffe://m/sa3"),
			[]*common_api.Match{exactSpiffe("spiffe://m/sa1"), exactSpiffe("spiffe://m/sa2"), prefixSpiffe("spiffe://m/")},
			`connection.uri_san_peer_certificate == "spiffe://m/sa3" && !(connection.uri_san_peer_certificate.startsWith("spiffe://m/"))`),
	)
})
