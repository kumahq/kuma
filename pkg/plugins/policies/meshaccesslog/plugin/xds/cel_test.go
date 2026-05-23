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
	sni := func(v string) *common_api.Match {
		return &common_api.Match{
			SNI: &common_api.SNIMatch{Type: common_api.SNIExactMatchType, Value: v},
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
		Entry("self with one prior",
			sni("sni-2"), []*common_api.Match{exactSpiffe("spiffe://m/sa1")},
			`connection.requested_server_name == "sni-2" && !(connection.uri_san_peer_certificate == "spiffe://m/sa1")`),
		Entry("self with two priors",
			sni("sni-3"),
			[]*common_api.Match{exactSpiffe("spiffe://m/sa1"), sni("sni-2")},
			`connection.requested_server_name == "sni-3" && !(connection.uri_san_peer_certificate == "spiffe://m/sa1") && !(connection.requested_server_name == "sni-2")`),
		Entry("nil prior is skipped defensively",
			sni("sni-x"),
			[]*common_api.Match{nil},
			`connection.requested_server_name == "sni-x"`),
	)
})
