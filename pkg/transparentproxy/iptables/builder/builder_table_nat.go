package builder

import (
	"fmt"
	"net"
	"strings"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/chains"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/tables"
)

func buildMeshInbound(
	cfg config.TrafficFlow,
	prefix string,
	meshInboundRedirect string,
) (*Chain, error) {
	meshInbound, err := NewChain(TableNat, cfg.Chain.GetFullName(prefix))
	if err != nil {
		return nil, err
	}

	if !cfg.Enabled {
		meshInbound.AddRule(
			Protocol(Tcp()),
			Jump(Return()),
		)
		return meshInbound, nil
	}

	// Include inbound ports
	for _, port := range cfg.IncludePorts {
		meshInbound.AddRule(
			Protocol(Tcp(DestinationPort(port))),
			Jump(ToUserDefinedChain(meshInboundRedirect)),
		)
	}

	if len(cfg.IncludePorts) == 0 {
		// Excluded outbound ports
		for _, port := range cfg.ExcludePorts {
			meshInbound.AddRule(
				Protocol(Tcp(DestinationPort(port))),
				Jump(Return()),
			)
		}
		meshInbound.AddRule(
			Protocol(Tcp()),
			Jump(ToUserDefinedChain(meshInboundRedirect)),
		)
	}

	return meshInbound, nil
}

func buildMeshOutbound(
	cfg config.InitializedConfig,
	dnsServers []string,
	ipv6 bool,
) (*Chain, error) {
	prefix := cfg.Redirect.NamePrefix
	inboundRedirectChainName := cfg.Redirect.Inbound.RedirectChain.GetFullName(prefix)
	outboundChainName := cfg.Redirect.Outbound.Chain.GetFullName(prefix)
	outboundRedirectChainName := cfg.Redirect.Outbound.RedirectChain.GetFullName(prefix)
	excludePorts := cfg.Redirect.Outbound.ExcludePorts
	includePorts := cfg.Redirect.Outbound.IncludePorts
	hasIncludedPorts := len(includePorts) > 0
	dnsRedirectPort := cfg.Redirect.DNS.Port
	uid := cfg.Owner.UID

	localhost := LocalhostCIDRIPv4
	inboundPassthroughSourceAddress := InboundPassthroughSourceAddressCIDRIPv4
	if ipv6 {
		inboundPassthroughSourceAddress = InboundPassthroughSourceAddressCIDRIPv6
		localhost = LocalhostCIDRIPv6
	}

	meshOutbound, err := NewChain(TableNat, outboundChainName)
	if err != nil {
		return nil, err
	}

	if !cfg.Redirect.Outbound.Enabled {
		meshOutbound.AddRule(
			Protocol(Tcp()),
			Jump(Return()),
		)
		return meshOutbound, nil
	}

	// Excluded outbound ports
	if !hasIncludedPorts {
		for _, port := range excludePorts {
			meshOutbound.AddRule(
				Protocol(Tcp(DestinationPort(port))),
				Jump(Return()),
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
		AddRule(
			Source(Address(inboundPassthroughSourceAddress)),
			OutInterface(cfg.LoopbackInterfaceName),
			Jump(Return()),
		).
		AddRule(
			Protocol(Tcp(NotDestinationPortIf(cfg.ShouldRedirectDNS, DNSPort))),
			OutInterface(cfg.LoopbackInterfaceName),
			NotDestination(localhost),
			Match(Owner(Uid(uid))),
			Jump(ToUserDefinedChain(inboundRedirectChainName)),
		).
		AddRule(
			Protocol(Tcp(NotDestinationPortIf(cfg.ShouldRedirectDNS, DNSPort))),
			OutInterface(cfg.LoopbackInterfaceName),
			Match(Owner(NotUid(uid))),
			Jump(Return()),
		).
		AddRule(
			Match(Owner(Uid(uid))),
			Jump(Return()),
		)
	if cfg.ShouldRedirectDNS() {
		if cfg.ShouldCaptureAllDNS() {
			meshOutbound.AddRule(
				Protocol(Tcp(DestinationPort(DNSPort))),
				Jump(ToPort(dnsRedirectPort)),
			)
		} else {
			for _, dnsIp := range dnsServers {
				meshOutbound.AddRule(
					Destination(dnsIp),
					Protocol(Tcp(DestinationPort(DNSPort))),
					Jump(ToPort(dnsRedirectPort)),
				)
			}
		}
	}
	meshOutbound.
		AddRule(
			Destination(localhost),
			Jump(Return()),
		)

	if hasIncludedPorts {
		for _, port := range includePorts {
			meshOutbound.AddRule(
				Protocol(Tcp(DestinationPort(port))),
				Jump(ToUserDefinedChain(outboundRedirectChainName)),
			)
		}
	} else {
		meshOutbound.AddRule(
			Jump(ToUserDefinedChain(outboundRedirectChainName)),
		)
	}

	return meshOutbound, nil
}

func buildMeshRedirect(cfg config.TrafficFlow, prefix string, ipv6 bool) (*Chain, error) {
	chainName := cfg.RedirectChain.GetFullName(prefix)

	redirectPort := cfg.Port
	if ipv6 && cfg.PortIPv6 != 0 {
		redirectPort = cfg.PortIPv6
	}

	redirectChain, err := NewChain(TableNat, chainName)
	if err != nil {
		return nil, err
	}

	return redirectChain.AddRule(
		Protocol(Tcp()),
		Jump(ToPort(redirectPort)),
	), nil
}

func addOutputRules(
	cfg config.InitializedConfig,
	dnsServers []string,
	nat *tables.NatTable,
	ipv6 bool,
) error {
	outboundChainName := cfg.Redirect.Outbound.Chain.GetFullName(cfg.Redirect.NamePrefix)
	dnsRedirectPort := cfg.Redirect.DNS.Port
	uid := cfg.Owner.UID
	rulePosition := uint(1)
	if cfg.Log.Enabled {
		nat.Output().AddRuleAtPosition(
			rulePosition,
			Jump(Log(OutputLogPrefix, cfg.Log.Level)),
		)
		rulePosition++
	}

	// Excluded outbound ports for UIDs
	for _, uIDsToPorts := range cfg.Redirect.Outbound.ExcludePortsForUIDs {
		var protocol *Parameter

		switch uIDsToPorts.Protocol {
		case TCP:
			protocol = Protocol(Tcp(DestinationPortRangeOrValue(uIDsToPorts)))
		case UDP:
			protocol = Protocol(Udp(DestinationPortRangeOrValue(uIDsToPorts)))
		default:
			return fmt.Errorf("unknown protocol %s, only 'tcp' or 'udp' allowed", uIDsToPorts.Protocol)
		}

		nat.Output().AddRuleAtPosition(
			rulePosition,
			Match(Multiport()),
			protocol,
			Match(Owner(UidRangeOrValue(uIDsToPorts))),
			Jump(Return()),
		)
		rulePosition++
	}

	if cfg.ShouldRedirectDNS() {
		jumpTarget := Return()
		if !ipv6 && cfg.ShouldFallbackDNSToUpstreamChain() {
			jumpTarget = ToUserDefinedChain(cfg.Redirect.DNS.UpstreamTargetChain)
		}

		nat.Output().AddRuleAtPosition(
			rulePosition,
			Protocol(Udp(DestinationPort(DNSPort))),
			Match(Owner(Uid(uid))),
			Jump(jumpTarget),
		)
		rulePosition++

		if cfg.ShouldCaptureAllDNS() {
			nat.Output().AddRuleAtPosition(
				rulePosition,
				Protocol(Udp(DestinationPort(DNSPort))),
				Jump(ToPort(dnsRedirectPort)),
			)
		} else {
			for _, dnsIp := range dnsServers {
				nat.Output().AddRuleAtPosition(
					rulePosition,
					Destination(dnsIp),
					Protocol(Udp(DestinationPort(DNSPort))),
					Jump(ToPort(dnsRedirectPort)),
				)
				rulePosition++
			}
		}
	}
	nat.Output().
		AddRule(
			Protocol(Tcp()),
			Jump(ToUserDefinedChain(outboundChainName)),
		)
	return nil
}

func addPreroutingRules(cfg config.InitializedConfig, nat *tables.NatTable, ipv6 bool) error {
	inboundChainName := cfg.Redirect.Inbound.Chain.GetFullName(cfg.Redirect.NamePrefix)
	rulePosition := uint(1)
	if cfg.Log.Enabled {
		nat.Prerouting().AddRule(
			Jump(Log(PreroutingLogPrefix, cfg.Log.Level)),
		)
	}

	if len(cfg.Redirect.VNet.Networks) > 0 {
		interfaceAndCidr := map[string]string{}
		for i := 0; i < len(cfg.Redirect.VNet.Networks); i++ {
			// we accept only first : so in case of IPv6 there should be no problem with parsing
			pair := strings.SplitN(cfg.Redirect.VNet.Networks[i], ":", 2)
			if len(pair) < 2 {
				return fmt.Errorf("incorrect definition of virtual network: %s", cfg.Redirect.VNet.Networks[i])
			}
			ipAddress, _, err := net.ParseCIDR(pair[1])
			if err != nil {
				return fmt.Errorf("incorrect CIDR definition: %s", err)
			}
			// if is ipv6 and address is ipv6 or is ipv4 and address is ipv4
			if (ipv6 && ipAddress.To4() == nil) || (!ipv6 && ipAddress.To4() != nil) {
				interfaceAndCidr[pair[0]] = pair[1]
			}
		}
		for iface, cidr := range interfaceAndCidr {
			nat.Prerouting().AddRuleAtPosition(
				rulePosition,
				InInterface(iface),
				Match(MatchUdp()),
				Protocol(Udp(DestinationPort(DNSPort))),
				Jump(ToPort(cfg.Redirect.DNS.Port)),
			)
			rulePosition += 1
			nat.Prerouting().AddRuleAtPosition(
				rulePosition,
				NotDestination(cidr),
				InInterface(iface),
				Protocol(Tcp()),
				Jump(ToPort(cfg.Redirect.Outbound.Port)),
			)
			rulePosition += 1
		}
		nat.Prerouting().AddRuleAtPosition(
			rulePosition,
			Protocol(Tcp()),
			Jump(ToUserDefinedChain(inboundChainName)),
		)
	} else {
		nat.Prerouting().AddRule(
			Protocol(Tcp()),
			Jump(ToUserDefinedChain(inboundChainName)),
		)
	}
	return nil
}

func buildNatTable(
	cfg config.InitializedConfig,
	dnsServers []string,
	ipv6 bool,
) (*tables.NatTable, error) {
	prefix := cfg.Redirect.NamePrefix
	inboundRedirectChainName := cfg.Redirect.Inbound.RedirectChain.GetFullName(prefix)

	nat := tables.Nat()

	if err := addOutputRules(cfg, dnsServers, nat, ipv6); err != nil {
		return nil, fmt.Errorf("could not add output rules %s", err)
	}
	if err := addPreroutingRules(cfg, nat, ipv6); err != nil {
		return nil, fmt.Errorf("could not add prerouting rules %s", err)
	}

	// MESH_INBOUND
	meshInbound, err := buildMeshInbound(cfg.Redirect.Inbound, prefix, inboundRedirectChainName)
	if err != nil {
		return nil, err
	}

	// MESH_INBOUND_REDIRECT
	meshInboundRedirect, err := buildMeshRedirect(cfg.Redirect.Inbound, prefix, ipv6)
	if err != nil {
		return nil, err
	}

	// MESH_OUTBOUND
	meshOutbound, err := buildMeshOutbound(cfg, dnsServers, ipv6)
	if err != nil {
		return nil, err
	}

	// MESH_OUTBOUND_REDIRECT
	meshOutboundRedirect, err := buildMeshRedirect(cfg.Redirect.Outbound, prefix, ipv6)
	if err != nil {
		return nil, err
	}

	return nat.
		WithCustomChain(meshInbound).
		WithCustomChain(meshOutbound).
		WithCustomChain(meshInboundRedirect).
		WithCustomChain(meshOutboundRedirect), nil
}
