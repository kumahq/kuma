package builder

import (
	"fmt"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/chains"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/rules"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/tables"
	"github.com/kumahq/kuma/pkg/util/maps"
)

func buildMeshInbound(cfg config.InitializedTrafficFlow) *Chain {
	meshInbound := MustNewChain(TableNat, cfg.ChainName)

	if !cfg.Enabled {
		return meshInbound.AddRules(
			rules.
				NewRule(
					Protocol(Tcp()),
					Jump(Return()),
				).
				WithComment("inbound traffic redirection is disabled"),
		)
	}

	for _, port := range cfg.IncludePorts {
		meshInbound.AddRules(
			rules.
				NewRule(
					Protocol(Tcp(DestinationPort(port))),
					Jump(ToUserDefinedChain(cfg.RedirectChainName)),
				).
				WithCommentf("redirect inbound traffic from port %d to the custom chain for processing", port),
		)
	}

	if len(cfg.IncludePorts) == 0 {
		for _, port := range cfg.ExcludePorts {
			meshInbound.AddRules(
				rules.
					NewRule(
						Protocol(Tcp(DestinationPort(port))),
						Jump(Return()),
					).
					WithCommentf("exclude inbound traffic from port %d from redirection", port),
			)
		}

		meshInbound.AddRules(
			rules.
				NewRule(
					Protocol(Tcp()),
					Jump(ToUserDefinedChain(cfg.RedirectChainName)),
				).
				WithComment("redirect all inbound traffic to the custom chain for processing"),
		)
	}

	return meshInbound
}

func buildMeshOutbound(cfg config.InitializedConfigIPvX) *Chain {
	meshOutbound := MustNewChain(TableNat, cfg.Redirect.Outbound.ChainName)

	if !cfg.Redirect.Outbound.Enabled {
		return meshOutbound.AddRules(
			rules.
				NewRule(
					Protocol(Tcp()),
					Jump(Return()),
				).
				WithComment("outbound traffic redirection is disabled"),
		)
	}

	if len(cfg.Redirect.Outbound.IncludePorts) == 0 {
		for _, port := range cfg.Redirect.Outbound.ExcludePorts {
			meshOutbound.AddRules(
				rules.
					NewRule(
						Protocol(Tcp(DestinationPort(port))),
						Jump(Return()),
					).
					WithCommentf("exclude outbound traffic from port %d from redirection", port),
			)
		}
	}

	meshOutbound.
		// ipv4:
		//   when tcp_packet to 192.168.0.10:7777 arrives ⤸
		//   iptables#nat ⤸
		//     PREROUTING ⤸
		//     MESH_INBOUND ⤸
		//     MESH_INBOUND_REDIRECT ⤸
		//   envoy@15006 ⤸
		//     listener#inbound:passthrough:ipv4 ⤸
		//     cluster#inbound:passthrough:ipv4 (source_ip 127.0.0.6) ⤸
		//     listener#192.168.0.10:7777 ⤸
		//     cluster#localhost:7777 ⤸
		//   localhost:7777
		//
		// ipv6:
		//   when tcp_packet to [fd00::0:10]:7777 arrives ⤸
		//   ip6tables#nat ⤸
		//     PREROUTING ⤸
		//     MESH_INBOUND ⤸
		//     MESH_INBOUND_REDIRECT ⤸
		//   envoy@15006 ⤸
		//     listener#inbound:passthrough:ipv6 ⤸
		//     cluster#inbound:passthrough:ipv6 (source_ip ::6) ⤸
		//     listener#[fd00::0:10]:7777 ⤸
		//     cluster#localhost:7777 ⤸
		//   localhost:7777
		AddRules(
			rules.
				NewRule(
					Source(Address(cfg.InboundPassthroughCIDR)),
					OutInterface(cfg.LoopbackInterfaceName),
					Jump(Return()),
				).
				WithCommentf("prevent traffic loops by ensuring traffic from the sidecar proxy (using %s) to loopback interface is not redirected again", cfg.InboundPassthroughCIDR),
			rules.
				NewRule(
					Protocol(Tcp(NotDestinationPortIfBool(cfg.Redirect.DNS.Enabled, DNSPort))),
					OutInterface(cfg.LoopbackInterfaceName),
					NotDestination(cfg.LocalhostCIDR),
					Match(Owner(Uid(cfg.Owner.UID))),
					Jump(ToUserDefinedChain(cfg.Redirect.Inbound.RedirectChainName)),
				).
				WithCommentf("redirect outbound TCP traffic (except to DNS port %d) destined for loopback interface, but not targeting address %s, and owned by UID %s (kuma-dp user) to %s chain for proper handling", DNSPort, cfg.LocalhostCIDR, cfg.Owner.UID, cfg.Redirect.Inbound.RedirectChainName),
			rules.
				NewRule(
					Protocol(Tcp(NotDestinationPortIfBool(cfg.Redirect.DNS.Enabled, DNSPort))),
					OutInterface(cfg.LoopbackInterfaceName),
					Match(Owner(NotUid(cfg.Owner.UID))),
					Jump(Return()),
				).
				WithCommentf("return outbound TCP traffic (except to DNS port %d) destined for loopback interface, owned by any UID other than %s (kuma-dp user)", DNSPort, cfg.Owner.UID),
			rules.
				NewRule(
					Match(Owner(Uid(cfg.Owner.UID))),
					Jump(Return()),
				).
				WithCommentf("return outbound traffic owned by UID %s (kuma-dp user)", cfg.Owner.UID),
		)

	if cfg.Redirect.DNS.Enabled {
		if cfg.Redirect.DNS.CaptureAll {
			meshOutbound.AddRules(
				rules.
					NewRule(
						Protocol(Tcp(DestinationPort(DNSPort))),
						Jump(ToPort(cfg.Redirect.DNS.Port)),
					).
					WithCommentf("redirect all DNS requests sent via TCP to kuma-dp DNS proxy (listening on port %d)", cfg.Redirect.DNS.Port),
			)
		} else {
			for _, dnsIp := range cfg.Redirect.DNS.Servers {
				meshOutbound.AddRules(
					rules.
						NewRule(
							Destination(dnsIp),
							Protocol(Tcp(DestinationPort(DNSPort))),
							Jump(ToPort(cfg.Redirect.DNS.Port)),
						).
						WithCommentf("redirect DNS requests sent via TCP to %s to kuma-dp DNS proxy (listening on port %d)", dnsIp, cfg.Redirect.DNS.Port),
				)
			}
		}
	}

	meshOutbound.AddRules(
		rules.
			NewRule(
				Destination(cfg.LocalhostCIDR),
				Jump(Return()),
			).
			WithCommentf("return traffic destined for localhost (%s) to avoid redirection", cfg.LocalhostCIDR),
	)

	for _, port := range cfg.Redirect.Outbound.IncludePorts {
		meshOutbound.AddRules(
			rules.
				NewRule(
					Protocol(Tcp(DestinationPort(port))),
					Jump(ToUserDefinedChain(cfg.Redirect.Outbound.RedirectChainName)),
				).
				WithCommentf("redirect outbound TCP traffic to port %d to our custom chain for further processing", port),
		)
	}

	if len(cfg.Redirect.Outbound.IncludePorts) == 0 {
		meshOutbound.AddRules(
			rules.
				NewRule(
					Jump(ToUserDefinedChain(cfg.Redirect.Outbound.RedirectChainName)),
				).
				WithComment("redirect all other outbound traffic to our custom chain for further processing"),
		)
	}

	return meshOutbound
}

// buildMeshRedirect creates a chain in the NAT table to handle traffic redirection
// to a specified port. The chain will be configured to redirect TCP traffic to the
// provided port, which can be different for IPv4 and IPv6.
func buildMeshRedirect(cfg config.InitializedTrafficFlow) *Chain {
	return MustNewChain(TableNat, cfg.RedirectChainName).AddRules(
		rules.
			NewRule(
				Protocol(Tcp()),
				Jump(ToPort(cfg.Port)),
			).
			WithCommentf("redirect TCP traffic to envoy (port %d)", cfg.Port),
	)
}

func addOutputRules(cfg config.InitializedConfigIPvX, nat *tables.NatTable) {
	rulePosition := uint(1)

	if cfg.Log.Enabled {
		nat.Output().AddRules(
			rules.
				NewRule(Jump(Log(OutputLogPrefix, cfg.Log.Level))).
				WithPosition(rulePosition).
				WithComment("log matching packets using kernel logging"),
		)
		rulePosition++
	}

	// Loop through UID-specific excluded ports and add corresponding NAT rules.
	for _, uIDsToPorts := range cfg.Redirect.Outbound.ExcludePortsForUIDs {
		var protocol *Parameter

		// Determine the protocol type and set up the correct parameter.
		switch uIDsToPorts.Protocol {
		case TCP:
			protocol = Protocol(Tcp(DestinationPortRangeOrValue(uIDsToPorts)))
		case UDP:
			protocol = Protocol(Udp(DestinationPortRangeOrValue(uIDsToPorts)))
		default:
			// This was already validated during config initialization, so this
			// warning should never appear
			cfg.Logger.Warnf(
				"unknown protocol %s, only 'tcp' or 'udp' allowed",
				uIDsToPorts.Protocol,
			)
			continue
		}

		nat.Output().AddRules(
			rules.
				NewRule(
					Match(Multiport()),
					protocol,
					Match(Owner(UidRangeOrValue(uIDsToPorts))),
					Jump(Return()),
				).
				WithPosition(rulePosition).
				WithComment("skip further processing for configured ports and UIDs"),
		)
		rulePosition++
	}

	// Conditionally add DNS redirection rules if DNS redirection is enabled.
	if cfg.Redirect.DNS.Enabled {
		nat.Output().AddRules(
			rules.
				NewRule(
					Protocol(Udp(DestinationPort(DNSPort))),
					Match(Owner(Uid(cfg.Owner.UID))),
					JumpConditional(
						cfg.Executables.Functionality.Chains.DockerOutput, // if DOCKER_OUTPUT should be targeted
						ToUserDefinedChain(ChainDockerOutput),             // --jump DOCKER_OUTPUT
						Return(),                                          // else RETURN
					),
				).
				WithPosition(rulePosition).
				WithConditionalComment(
					cfg.Executables.Functionality.Chains.DockerOutput,
					fmt.Sprintf(
						"redirect DNS traffic from kuma-dp to the %s chain",
						ChainDockerOutput,
					),
					"return early for DNS traffic from kuma-dp",
				),
		)
		rulePosition++

		if cfg.Redirect.DNS.CaptureAll {
			nat.Output().AddRules(
				rules.
					NewRule(
						Protocol(Udp(DestinationPort(DNSPort))),
						Jump(ToPort(cfg.Redirect.DNS.Port)),
					).
					WithPosition(rulePosition).
					WithCommentf("redirect all DNS requests to the kuma-dp DNS proxy (listening on port %d)", cfg.Redirect.DNS.Port),
			)
		} else {
			for _, dnsIp := range cfg.Redirect.DNS.Servers {
				nat.Output().AddRules(
					rules.
						NewRule(
							Destination(dnsIp),
							Protocol(Udp(DestinationPort(DNSPort))),
							Jump(ToPort(cfg.Redirect.DNS.Port)),
						).
						WithPosition(rulePosition).
						WithCommentf("redirect DNS requests to %s to the kuma-dp DNS proxy (listening on port %d)", dnsIp, cfg.Redirect.DNS.Port),
				)
				rulePosition++
			}
		}
	}

	nat.Output().AddRules(
		rules.
			NewRule(
				Protocol(Tcp()),
				Jump(ToUserDefinedChain(cfg.Redirect.Outbound.ChainName)),
			).
			WithComment("redirect outbound TCP traffic to our custom chain for processing"),
	)
}

// addPreroutingRules adds rules to the PREROUTING chain of the NAT table to
// handle inbound traffic according to the provided configuration.
func addPreroutingRules(cfg config.InitializedConfigIPvX, nat *tables.NatTable) {
	rulePosition := uint(1)

	// Add a logging rule if logging is enabled.
	if cfg.Log.Enabled {
		nat.Prerouting().AddRules(
			rules.
				NewRule(Jump(Log(PreroutingLogPrefix, cfg.Log.Level))).
				WithComment("log matching packets using kernel logging"),
		)
	}

	if len(cfg.Redirect.VNet.InterfaceCIDRs) == 0 {
		nat.Prerouting().AddRules(
			rules.
				NewRule(
					Protocol(Tcp()),
					Jump(ToUserDefinedChain(cfg.Redirect.Inbound.ChainName)),
				).
				WithComment("redirect inbound TCP traffic to our custom chain for processing"),
		)
		return
	}

	for _, iface := range maps.SortedKeys(cfg.Redirect.VNet.InterfaceCIDRs) {
		nat.Prerouting().AddRules(
			rules.
				NewRule(
					InInterface(iface),
					Match(MatchUdp()),
					Protocol(Udp(DestinationPort(DNSPort))),
					Jump(ToPort(cfg.Redirect.DNS.Port)),
				).
				WithPosition(rulePosition).
				WithCommentf("redirect DNS requests on interface %s to the kuma-dp DNS proxy (listening on port %d)", iface, cfg.Redirect.DNS.Port),
			rules.
				NewRule(
					NotDestination(cfg.Redirect.VNet.InterfaceCIDRs[iface]),
					InInterface(iface),
					Protocol(Tcp()),
					Jump(ToPort(cfg.Redirect.Outbound.Port)),
				).
				WithPosition(rulePosition+1).
				WithCommentf("redirect TCP traffic on interface %s, excluding destination %s, to the envoy's outbound passthrough port %d", iface, cfg.Redirect.VNet.InterfaceCIDRs[iface], cfg.Redirect.Outbound.Port),
		)
		rulePosition += 2
	}

	nat.Prerouting().AddRules(
		rules.
			NewRule(
				Protocol(Tcp()),
				Jump(ToUserDefinedChain(cfg.Redirect.Inbound.ChainName)),
			).
			WithPosition(rulePosition).
			WithComment("redirect remaining TCP traffic to our custom chain for processing"),
	)
}

// buildNatTable constructs the NAT table for iptables with the necessary rules
// for handling inbound and outbound traffic redirection, DNS redirection, and
// specific port exclusions or inclusions. It sets up custom chains for mesh
// traffic management based on the provided configuration.
func buildNatTable(cfg config.InitializedConfigIPvX) *tables.NatTable {
	nat := tables.Nat()

	addOutputRules(cfg, nat)

	addPreroutingRules(cfg, nat)

	return nat.
		WithCustomChain(buildMeshInbound(cfg.Redirect.Inbound)).
		WithCustomChain(buildMeshOutbound(cfg)).
		WithCustomChain(buildMeshRedirect(cfg.Redirect.Inbound)).
		WithCustomChain(buildMeshRedirect(cfg.Redirect.Outbound))
}
