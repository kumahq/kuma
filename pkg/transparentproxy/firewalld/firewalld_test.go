package firewalld

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/test/matchers"
)

type testCase struct {
	iptablesRules map[string][]string
	goldenFile    string
}

var _ = Describe("firewalld", func() {

	DescribeTable("check xml genrate",
		func(given testCase) {
			translator := NewFirewalldIptablesTranslator(true)
			out, err := translator.StoreRules(given.iptablesRules)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(MatchGoldenXML("testdata", given.goldenFile))
		},
		Entry("should generate xml", testCase{
			iptablesRules: map[string][]string{
				"nat": {
					"-N KUMA_INBOUND",
					"-N KUMA_REDIRECT",
					"-N KUMA_IN_REDIRECT",
					"-N KUMA_OUTPUT",
					"-A KUMA_INBOUND -p tcp --dport 15008 -j RETURN",
					"-A KUMA_REDIRECT -p tcp -j REDIRECT --to-ports 15001",
					"-A KUMA_IN_REDIRECT -p tcp -j REDIRECT --to-ports 15006",
					"-A PREROUTING -p tcp -j KUMA_INBOUND",
					"-A KUMA_INBOUND -p tcp --dport 22 -j RETURN",
					"-A KUMA_INBOUND -p tcp -j KUMA_IN_REDIRECT",
					"-A OUTPUT -p tcp -j KUMA_OUTPUT",
					"-A KUMA_OUTPUT -o lo -s 127.0.0.6/32 -j RETURN",
					"-A KUMA_OUTPUT -o lo ! -d 127.0.0.1/32 -m owner --uid-owner 0 -j KUMA_IN_REDIRECT",
					"-A KUMA_OUTPUT -o lo -m owner ! --uid-owner 0 -j RETURN",
					"-A KUMA_OUTPUT -m owner --uid-owner 0 -j RETURN",
					"-A KUMA_OUTPUT -o lo ! -d 127.0.0.1/32 -m owner --gid-owner 0 -j KUMA_IN_REDIRECT",
					"-A KUMA_OUTPUT -o lo -m owner ! --gid-owner 0 -j RETURN",
					"-A KUMA_OUTPUT -m owner --gid-owner 0 -j RETURN",
					"-A KUMA_OUTPUT -d 127.0.0.1/32 -j RETURN",
					"-A KUMA_OUTPUT -j KUMA_REDIRECT",
				},
			},
			goldenFile: "full_direct.xml",
		}),
		Entry("should generate xml", testCase{
			iptablesRules: map[string][]string{
				"nat": {
					"-N KUMA_INBOUND",
					"-N KUMA_OUTPUT",
					"-A KUMA_INBOUND -p tcp --dport 15008 -j RETURN",
					"-A PREROUTING -p tcp -j KUMA_INBOUND",
					"-A KUMA_INBOUND -p tcp --dport 22 -j RETURN",
					"-A KUMA_INBOUND -p tcp --dport 22 -j RETURN",
					"-A KUMA_INBOUND -p tcp --dport 22 -j RETURN",
					"-A OUTPUT -p tcp -j KUMA_OUTPUT",
					"-A KUMA_OUTPUT -d 127.0.0.1/32 -j RETURN",
					"-A OUTPUT -p tcp -j KUMA_OUTPUT",
					"-A KUMA_OUTPUT -d 127.0.0.1/32 -j RETURN",
					"-N KUMA_OUTPUT",
					"-N KUMA_INBOUND",
				},
			},
			goldenFile: "no_duplicate_direct.xml",
		}),
	)

})
