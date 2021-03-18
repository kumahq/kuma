package firewalld

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/test/matchers"
)

/*

* nat
-N KUMA_INBOUND
-N KUMA_REDIRECT
-N KUMA_IN_REDIRECT
-N KUMA_OUTPUT
-A KUMA_INBOUND -p tcp --dport 15008 -j RETURN
-A KUMA_REDIRECT -p tcp -j REDIRECT --to-ports 15001
-A KUMA_IN_REDIRECT -p tcp -j REDIRECT --to-ports 15006
-A PREROUTING -p tcp -j KUMA_INBOUND
-A KUMA_INBOUND -p tcp --dport 22 -j RETURN
-A KUMA_INBOUND -p tcp -j KUMA_IN_REDIRECT
-A OUTPUT -p tcp -j KUMA_OUTPUT
-A KUMA_OUTPUT -o lo -s ::6/128 -j RETURN
-A KUMA_OUTPUT -o lo ! -d ::1/128 -m owner --uid-owner 0 -j KUMA_IN_REDIRECT
-A KUMA_OUTPUT -o lo -m owner ! --uid-owner 0 -j RETURN
-A KUMA_OUTPUT -m owner --uid-owner 0 -j RETURN
-A KUMA_OUTPUT -o lo ! -d ::1/128 -m owner --gid-owner 0 -j KUMA_IN_REDIRECT
-A KUMA_OUTPUT -o lo -m owner ! --gid-owner 0 -j RETURN
-A KUMA_OUTPUT -m owner --gid-owner 0 -j RETURN
-A KUMA_OUTPUT -d ::1/128 -j RETURN
-A KUMA_OUTPUT -j KUMA_REDIRECT
COMMIT

*/

var _ = Describe("firewalld", func() {
	var iptablesRules = map[string][]string{
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
	}

	It("parse a single rule", func() {
		translator := NewFirewalldIptablesTranslator(true)
		out, err := translator.StoreRules(iptablesRules)
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(MatchGoldenXML(filepath.Join("testdata", "full_direct.xml")))
	})
})
