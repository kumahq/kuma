package builder

import (
	"fmt"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/consts"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/chains"
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
				NewAppendRule(
					Protocol(Tcp()),
					Jump(Return()),
				).
				WithComment("inbound traffic redirection is disabled"),
		)
	}

	// TODO(bartsmykla): Consider combining this loop with the one processing
	//  `cfg.ExcludePorts` for logical consistency. However, this requires
	//  careful handling as parsing `cfg.ExcludePorts` as `Exclusion`s would
	//  alter the rule placement for **outbound**:
	//  - Currently, exclusion rules for `--exclude-outbound-ports` are placed
	//    in the KUMA_MESH_OUTBOUND chain.
	//  - Exclusions from `--exclude-outbound-ports-for-uids` are placed in the
	//    OUTPUT chain.
	//  Combining these may require revisiting how we structure rule generation
	//  for outbound traffic to maintain correct behavior.
	for _, exclusion := range cfg.Exclusions {
		meshInbound.AddRules(
			rules.
				NewAppendRule(
					Source(Address(exclusion.Address)),
					Jump(Return()),
				).
				WithComment("skip further processing for configured IP address"),
		)
	}

	for _, port := range cfg.IncludePorts {
		meshInbound.AddRules(
			rules.
				NewAppendRule(
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
					NewAppendRule(
						Protocol(Tcp(DestinationPort(port))),
						Jump(Return()),
					).
					WithCommentf("exclude inbound traffic from port %d from redirection", port),
			)
		}

		meshInbound.AddRules(
			rules.
				NewAppendRule(
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
				NewAppendRule(
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
					NewAppendRule(
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
				NewAppendRule(
					Source(Address(cfg.InboundPassthroughCIDR)),
					OutInterface(cfg.LoopbackInterfaceName),
					Jump(Return()),
				).
				WithCommentf("prevent traffic loops by ensuring traffic from the sidecar proxy (using %s) to loopback interface is not redirected again", cfg.InboundPassthroughCIDR),
			rules.
				NewAppendRule(
					Protocol(Tcp(NotDestinationPortIfBool(cfg.Redirect.DNS.Enabled, DNSPort))),
					OutInterface(cfg.LoopbackInterfaceName),
					NotDestination(cfg.LocalhostCIDR),
					Match(Owner(Uid(cfg.KumaDPUser.UID))),
					Jump(ToUserDefinedChain(cfg.Redirect.Inbound.RedirectChainName)),
				).
				WithCommentf("redirect outbound TCP traffic (except to DNS port %d) destined for loopback interface, but not targeting address %s, and owned by UID %s (kuma-dp user) to %s chain for proper handling", DNSPort, cfg.LocalhostCIDR, cfg.KumaDPUser.UID, cfg.Redirect.Inbound.RedirectChainName),
			rules.
				NewAppendRule(
					Protocol(Tcp(NotDestinationPortIfBool(cfg.Redirect.DNS.Enabled, DNSPort))),
					OutInterface(cfg.LoopbackInterfaceName),
					Match(Owner(NotUid(cfg.KumaDPUser.UID))),
					Jump(Return()),
				).
				WithCommentf("return outbound TCP traffic (except to DNS port %d) destined for loopback interface, owned by any UID other than %s (kuma-dp user)", DNSPort, cfg.KumaDPUser.UID),
			rules.
				NewAppendRule(
					Match(Owner(Uid(cfg.KumaDPUser.UID))),
					Jump(Return()),
				).
				WithCommentf("return outbound traffic owned by UID %s (kuma-dp user)", cfg.KumaDPUser.UID),
		)

	if cfg.Redirect.DNS.Enabled {
		if cfg.Redirect.DNS.CaptureAll {
			meshOutbound.AddRules(
				rules.
					NewAppendRule(
						Protocol(Tcp(DestinationPort(DNSPort))),
						Jump(ToPort(cfg.Redirect.DNS.Port)),
					).
					WithCommentf("redirect all DNS requests sent via TCP to kuma-dp DNS proxy (listening on port %d)", cfg.Redirect.DNS.Port),
			)
		} else {
			for _, dnsIp := range cfg.Redirect.DNS.Servers {
				meshOutbound.AddRules(
					rules.
						NewAppendRule(
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
			NewAppendRule(
				Destination(cfg.LocalhostCIDR),
				Jump(Return()),
			).
			WithCommentf("return traffic destined for localhost (%s) to avoid redirection", cfg.LocalhostCIDR),
	)

	for _, port := range cfg.Redirect.Outbound.IncludePorts {
		meshOutbound.AddRules(
			rules.
				NewAppendRule(
					Protocol(Tcp(DestinationPort(port))),
					Jump(ToUserDefinedChain(cfg.Redirect.Outbound.RedirectChainName)),
				).
				WithCommentf("redirect outbound TCP traffic to port %d to our custom chain for further processing", port),
		)
	}

	if len(cfg.Redirect.Outbound.IncludePorts) == 0 {
		meshOutbound.AddRules(
			rules.
				NewAppendRule(
					Jump(ToUserDefinedChain(cfg.Redirect.Outbound.RedirectChainName)),
				).
				WithComment("redirect all other outbound traffic to our custom chain for further processing"),
		)
	}

	return meshOutbound
}

// buildMeshRedirect creates a chain in the NAT table to handle traffic redirection
// to a specified port. The chain will be configured to redirect TCP traffic to the
// provided port, which can be different for IPv4 and IPv6
func buildMeshRedirect(cfg config.InitializedTrafficFlow) *Chain {
	return MustNewChain(TableNat, cfg.RedirectChainName).AddRules(
		rules.
			NewAppendRule(
				Protocol(Tcp()),
				Jump(ToPort(cfg.Port)),
			).
			WithCommentf("redirect TCP traffic to envoy (port %d)", cfg.Port),
	)
}

func addOutputRules(cfg config.InitializedConfigIPvX, nat *tables.NatTable) {
	if cfg.Log.Enabled {
		nat.Output().AddRules(
			rules.
				NewInsertRule(Jump(Log(OutputLogPrefix, cfg.Log.Level))).
				WithComment("log matching packets using kernel logging"),
		)
	}

	for _, exclusion := range cfg.Redirect.Outbound.Exclusions {
		nat.Output().AddRules(
			rules.
				NewInsertRule(
					MatchIf(exclusion.Ports != "", Multiport()),
					Protocol(
						TcpIf(
							exclusion.Protocol == ProtocolTCP,
							DestinationPortRangeOrValue(exclusion),
						),
						UdpIf(
							exclusion.Protocol == ProtocolUDP,
							DestinationPortRangeOrValue(exclusion),
						),
					),
					MatchIf(
						exclusion.UIDs != "",
						Owner(UidRangeOrValue(exclusion)),
					),
					Destination(exclusion.Address),
					Jump(Return()),
				).
				WithComment("skip further processing for configured IP addresses, ports and UIDs"),
		)
	}

	if cfg.Redirect.DNS.Enabled {
		nat.Output().AddRules(
			rules.
				NewInsertRule(
					Protocol(Udp(DestinationPort(DNSPort))),
					Match(Owner(Uid(cfg.KumaDPUser.UID))),
					JumpConditional(
						// if DOCKER_OUTPUT should be targeted --jump DOCKER_OUTPUT or else RETURN
						cfg.Executables.Functionality.Chains.DockerOutput,
						ToUserDefinedChain(ChainDockerOutput),
						Return(),
					),
				).
				WithConditionalComment(
					cfg.Executables.Functionality.Chains.DockerOutput,
					fmt.Sprintf(
						"redirect DNS traffic from kuma-dp to the %s chain",
						ChainDockerOutput,
					),
					"return early for DNS traffic from kuma-dp",
				),
		)

		if cfg.Redirect.DNS.CaptureAll {
			nat.Output().AddRules(
				rules.
					NewInsertRule(
						Protocol(Udp(DestinationPort(DNSPort))),
						Jump(ToPort(cfg.Redirect.DNS.Port)),
					).
					WithCommentf("redirect all DNS requests to the kuma-dp DNS proxy (listening on port %d)", cfg.Redirect.DNS.Port),
			)
		} else {
			for _, dnsIp := range cfg.Redirect.DNS.Servers {
				nat.Output().AddRules(
					rules.
						NewInsertRule(
							Destination(dnsIp),
							Protocol(Udp(DestinationPort(DNSPort))),
							Jump(ToPort(cfg.Redirect.DNS.Port)),
						).
						WithCommentf("redirect DNS requests to %s to the kuma-dp DNS proxy (listening on port %d)", dnsIp, cfg.Redirect.DNS.Port),
				)
			}
		}
	}

	nat.Output().AddRules(
		rules.
			NewConditionalInsertOrAppendRule(
				cfg.Redirect.Outbound.InsertRedirectInsteadOfAppend,
				Protocol(Tcp()),
				Jump(ToUserDefinedChain(cfg.Redirect.Outbound.ChainName)),
			).
			WithComment("redirect outbound TCP traffic to our custom chain for processing"),
	)
}

// addPreroutingRules adds rules to the PREROUTING chain of the NAT table to
// handle inbound traffic according to the provided configuration
func addPreroutingRules(cfg config.InitializedConfigIPvX, nat *tables.NatTable) {
	// Add a logging rule if logging is enabled.
	if cfg.Log.Enabled {
		nat.Prerouting().AddRules(
			rules.
				NewAppendRule(Jump(Log(PreroutingLogPrefix, cfg.Log.Level))).
				WithComment("log matching packets using kernel logging"),
		)
	}

	if len(cfg.Redirect.VNet.InterfaceCIDRs) == 0 && !cfg.Redirect.Inbound.InsertRedirectInsteadOfAppend {
		nat.Prerouting().AddRules(
			rules.
				NewAppendRule(
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
				NewInsertRule(
					InInterface(iface),
					Match(MatchUdp()),
					Protocol(Udp(DestinationPort(DNSPort))),
					Jump(ToPort(cfg.Redirect.DNS.Port)),
				).
				WithCommentf("redirect DNS requests on interface %s to the kuma-dp DNS proxy (listening on port %d)", iface, cfg.Redirect.DNS.Port),
			rules.
				NewInsertRule(
					NotDestination(cfg.Redirect.VNet.InterfaceCIDRs[iface]),
					InInterface(iface),
					Protocol(Tcp()),
					Jump(ToPort(cfg.Redirect.Outbound.Port)),
				).
				WithCommentf("redirect TCP traffic on interface %s, excluding destination %s, to the envoy's outbound passthrough port %d", iface, cfg.Redirect.VNet.InterfaceCIDRs[iface], cfg.Redirect.Outbound.Port),
		)
	}

	nat.Prerouting().AddRules(
		rules.
			NewInsertRule(
				Protocol(Tcp()),
				Jump(ToUserDefinedChain(cfg.Redirect.Inbound.ChainName)),
			).
			WithComment("redirect remaining TCP traffic to our custom chain for processing"),
	)
}

// buildNatTable constructs the NAT table for iptables with the necessary rules
// for handling inbound and outbound traffic redirection, DNS redirection, and
// specific port exclusions or inclusions. It sets up custom chains for mesh
// traffic management based on the provided configuration
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
