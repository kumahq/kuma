package xds_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	. "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/plugin/xds"
)

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
