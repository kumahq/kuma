package builder

import (
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/consts"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/tables"
	tproxy_test "github.com/kumahq/kuma/pkg/transparentproxy/test"
)

var _ = Describe("Builder nat", func() {
	DescribeTable("should insert PREROUTING rules",
		func(vnet []string, verbose bool, ipv6 bool, expect ...string) {
			// given
			nat := tables.Nat()
			cfg := tproxy_test.InitializeConfigIPvX(config.Config{
				Redirect: config.Redirect{
					NamePrefix: consts.IptablesChainsPrefix,
					Inbound: config.TrafficFlow{
						Enabled:           true,
						Port:              1234,
						ChainName:         "INBOUND",
						RedirectChainName: "INBOUND_REDIRECT",
					},
					Outbound: config.TrafficFlow{
						Enabled:           true,
						Port:              12345,
						ChainName:         "OUTBOUND",
						RedirectChainName: "OUTBOUND_REDIRECT",
					},
					DNS: config.DNS{Port: 15053},
					VNet: config.VNet{
						Networks: vnet,
					},
				},
				Verbose: verbose,
			}, ipv6)

			// when
			addPreroutingRules(cfg, nat)
			table := tables.BuildRulesForRestore(cfg, nat)

			// then
			for _, rule := range expect {
				Expect(table).To(WithTransform(func(in string) string {
					// Remove comments
					return regexp.
						MustCompile(`(?:-m|--match) comment --comment ".*?" `).
						ReplaceAllString(in, "")
				}, ContainSubstring(rule)))
			}
		},
		Entry("ipv4 not verbose",
			[]string{"docker:1.2.3.4/24", "br+:127.0.0.0/32"},
			false,
			false,
			// rules have random order, so we cannot compare addresses and names
			"-I PREROUTING 1",
			"-i docker -m udp -p udp --dport 53 -j REDIRECT --to-ports 15053",
			"-I PREROUTING 2",
			"! -d 1.2.3.4/24 -i docker -p tcp -j REDIRECT --to-ports 12345",
			"-I PREROUTING 3",
			"-i br+ -m udp -p udp --dport 53 -j REDIRECT --to-ports 15053",
			"-I PREROUTING 4",
			"! -d 127.0.0.0/32 -i br+ -p tcp -j REDIRECT --to-ports 12345",
			"-I PREROUTING 5 -p tcp -j KUMA_MESH_INBOUND",
		),
		Entry("ipv4 not verbose",
			[]string{"docker:1.2.3.4/24", "br+:127.0.0.0/32"},
			true,
			false,
			"--insert PREROUTING 1",
			"--in-interface docker --match udp --protocol udp --destination-port 53 --jump REDIRECT --to-ports 15053",
			"--insert PREROUTING 2",
			"! --destination 1.2.3.4/24 --in-interface docker --protocol tcp --jump REDIRECT --to-ports 12345",
			"--insert PREROUTING 3",
			"--in-interface br+ --match udp --protocol udp --destination-port 53 --jump REDIRECT --to-ports 15053",
			"--insert PREROUTING 4",
			"! --destination 127.0.0.0/32 --in-interface br+ --protocol tcp --jump REDIRECT --to-ports 12345",
			"--insert PREROUTING 5 --protocol tcp --jump KUMA_MESH_INBOUND",
		),
		Entry("ipv6 not verbose",
			[]string{"docker:::6/24", "br+:1::1/128"},
			false,
			true,
			"-I PREROUTING 1",
			"-i docker -m udp -p udp --dport 53 -j REDIRECT --to-ports 15053",
			"-I PREROUTING 2",
			"! -d ::6/24 -i docker -p tcp -j REDIRECT --to-ports 12345",
			"-I PREROUTING 3",
			"-i br+ -m udp -p udp --dport 53 -j REDIRECT --to-ports 15053",
			"-I PREROUTING 4",
			"! -d 1::1/128 -i br+ -p tcp -j REDIRECT --to-ports 12345",
			"-I PREROUTING 5 -p tcp -j KUMA_MESH_INBOUND",
		),
		Entry("ipv6 not verbose",
			[]string{"docker:::6/24", "br+:1::1/128"},
			true,
			true,
			"--insert PREROUTING 1",
			"--in-interface docker --match udp --protocol udp --destination-port 53 --jump REDIRECT --to-ports 15053",
			"--insert PREROUTING 2",
			"! --destination ::6/24 --in-interface docker --protocol tcp --jump REDIRECT --to-ports 12345",
			"--insert PREROUTING 3",
			"--in-interface br+ --match udp --protocol udp --destination-port 53 --jump REDIRECT --to-ports 15053",
			"--insert PREROUTING 4",
			"! --destination 1::1/128 --in-interface br+ --protocol tcp --jump REDIRECT --to-ports 12345",
			"--insert PREROUTING 5 --protocol tcp --jump KUMA_MESH_INBOUND",
		),
		Entry("ipv4 without ipv6 rules",
			[]string{"docker:127.0.0.6/24", "br+:1::1/128"},
			true,
			false,
			"--insert PREROUTING 1 --in-interface docker --match udp --protocol udp --destination-port 53 --jump REDIRECT --to-ports 15053",
			"--insert PREROUTING 2 ! --destination 127.0.0.6/24 --in-interface docker --protocol tcp --jump REDIRECT --to-ports 12345",
			"--insert PREROUTING 3 --protocol tcp --jump KUMA_MESH_INBOUND",
		),
		Entry("ipv6 without ipv4 rules",
			[]string{"docker:127.0.0.6/24", "br+:1::1/128"},
			true,
			true,
			"--insert PREROUTING 1 --in-interface br+ --match udp --protocol udp --destination-port 53 --jump REDIRECT --to-ports 15053",
			"--insert PREROUTING 2 ! --destination 1::1/128 --in-interface br+ --protocol tcp --jump REDIRECT --to-ports 12345",
			"--insert PREROUTING 3 --protocol tcp --jump KUMA_MESH_INBOUND",
		),
	)

	DescribeTable("should append PREROUTING rules",
		func(verbose bool, ipv6 bool, expect ...string) {
			// given
			nat := tables.Nat()
			cfg := tproxy_test.InitializeConfigIPvX(config.Config{
				Redirect: config.Redirect{
					NamePrefix: consts.IptablesChainsPrefix,
					Inbound: config.TrafficFlow{
						Enabled:           true,
						Port:              1234,
						ChainName:         "INBOUND",
						RedirectChainName: "INBOUND_REDIRECT",
					},
					Outbound: config.TrafficFlow{
						Enabled:           true,
						Port:              12345,
						ChainName:         "OUTBOUND",
						RedirectChainName: "OUTBOUND_REDIRECT",
					},
					DNS: config.DNS{Port: 15053},
				},
			}, ipv6)

			// when
			addPreroutingRules(cfg, nat)
			table := tables.BuildRulesForRestore(cfg, nat)

			// then
			for _, rule := range expect {
				Expect(table).To(WithTransform(func(in string) string {
					// Remove comments
					return regexp.
						MustCompile(`(?:-m|--match) comment --comment ".*?" `).
						ReplaceAllString(in, "")
				}, ContainSubstring(rule)))
			}
		},
		Entry("ipv4 not verbose", false, false, "-A PREROUTING -p tcp -j KUMA_MESH_INBOUND"),
	)
})
