package builder

import (
	"fmt"
	"net"
	"strings"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/chains"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/rules"
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
		meshInbound.AddRules(
			rules.
				NewRule(
					Protocol(Tcp()),
					Jump(Return()),
				).
				WithComment("inbound traffic redirection is disabled"),
		)
		return meshInbound, nil
	}

	// Include specific inbound ports for redirection
	for _, port := range cfg.IncludePorts {
		meshInbound.AddRules(
			rules.
				NewRule(
					Protocol(Tcp(DestinationPort(port))),
					Jump(ToUserDefinedChain(meshInboundRedirect)),
				).
				WithCommentf("redirect inbound traffic from port %d to the custom chain for processing", port),
		)
	}

	if len(cfg.IncludePorts) == 0 {
		// Exclude specific inbound ports from redirection
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

		// Redirect all other inbound traffic to the mesh inbound redirect chain
		meshInbound.AddRules(
			rules.
				NewRule(
					Protocol(Tcp()),
					Jump(ToUserDefinedChain(meshInboundRedirect)),
				).
				WithComment("redirect all inbound traffic to the custom chain for processing"),
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
	shouldRedirectDNS := cfg.Redirect.DNS.EnabledIPv4
	if ipv6 {
		inboundPassthroughSourceAddress = InboundPassthroughSourceAddressCIDRIPv6
		localhost = LocalhostCIDRIPv6
		shouldRedirectDNS = cfg.Redirect.DNS.EnabledIPv6
	}

	meshOutbound, err := NewChain(TableNat, outboundChainName)
	if err != nil {
		return nil, err
	}

	if !cfg.Redirect.Outbound.Enabled {
		meshOutbound.AddRules(
			rules.
				NewRule(
					Protocol(Tcp()),
					Jump(Return()),
				).
				WithComment("outbound traffic redirection is disabled"),
		)
		return meshOutbound, nil
	}

	// Exclude traffic from redirection on specified outbound ports
	if !hasIncludedPorts {
		for _, port := range excludePorts {
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
					Source(Address(inboundPassthroughSourceAddress)),
					OutInterface(cfg.LoopbackInterfaceName),
					Jump(Return()),
				).
				WithCommentf("prevent traffic loops by ensuring traffic from the sidecar proxy (using %s) to loopback interface is not redirected again", inboundPassthroughSourceAddress),
			// Redirect outbound TCP traffic (except DNS port 53) from the
			// loopback interface, not targeting localhost, and owned by UID
			// 5678 to the inbound redirect chain for proper handling by mesh's
			// inbound rules.
			rules.
				NewRule(
					Protocol(Tcp(NotDestinationPortIfBool(shouldRedirectDNS, DNSPort))),
					OutInterface(cfg.LoopbackInterfaceName),
					NotDestination(localhost),
					Match(Owner(Uid(uid))),
					Jump(ToUserDefinedChain(inboundRedirectChainName)),
				).
				WithCommentf("redirect outbound TCP traffic (except to DNS port %d) destined for loopback interface, but not targeting address %s, and owned by UID %s (kuma-dp user) to %s chain for proper handling", DNSPort, localhost, uid, inboundRedirectChainName),
			// Return outbound TCP traffic (except DNS port 53) from loopback
			// interface, owned by any UID other than 5678.
			rules.
				NewRule(
					Protocol(Tcp(NotDestinationPortIfBool(shouldRedirectDNS, DNSPort))),
					OutInterface(cfg.LoopbackInterfaceName),
					Match(Owner(NotUid(uid))),
					Jump(Return()),
				).
				WithCommentf("return outbound TCP traffic (except to DNS port %d) destined for loopback interface, owned by any UID other than %s (kuma-dp user)", DNSPort, uid),
			// Return outbound traffic owned by UID 5678.
			rules.
				NewRule(
					Match(Owner(Uid(uid))),
					Jump(Return()),
				).
				WithCommentf("return outbound traffic owned by UID %s (kuma-dp user)", uid),
		)

	if shouldRedirectDNS {
		if cfg.ShouldCaptureAllDNS() {
			// Redirect all DNS requests to kuma-dp DNS proxy.
			meshOutbound.AddRules(
				rules.
					NewRule(
						Protocol(Tcp(DestinationPort(DNSPort))),
						Jump(ToPort(dnsRedirectPort)),
					).
					WithCommentf("redirect all DNS requests sent via TCP to kuma-dp DNS proxy (listening on port %d)", dnsRedirectPort),
			)
		} else {
			for _, dnsIp := range dnsServers {
				// Redirect DNS requests to specified DNS resolvers to kuma-dp DNS proxy.
				meshOutbound.AddRules(
					rules.
						NewRule(
							Destination(dnsIp),
							Protocol(Tcp(DestinationPort(DNSPort))),
							Jump(ToPort(dnsRedirectPort)),
						).
						WithCommentf("redirect DNS requests sent via TCP to %s to kuma-dp DNS proxy (listening on port %d)", dnsIp, dnsRedirectPort),
				)
			}
		}
	}

	meshOutbound.AddRules(
		// Return traffic destined for localhost to avoid redirection.
		rules.
			NewRule(
				Destination(localhost),
				Jump(Return()),
			).
			WithCommentf("return traffic destined for localhost (%s) to avoid redirection", localhost),
	)

	if hasIncludedPorts {
		for _, port := range includePorts {
			// Redirect outbound traffic from the specified included ports to
			// the custom outbound redirect chain.
			meshOutbound.AddRules(
				rules.
					NewRule(
						Protocol(Tcp(DestinationPort(port))),
						Jump(ToUserDefinedChain(outboundRedirectChainName)),
					).
					WithCommentf("redirect outbound TCP traffic to port %d to our custom chain for further processing", port),
			)
		}
	} else {
		// Redirect all other outbound traffic to the custom outbound redirect
		// chain.
		meshOutbound.AddRules(
			rules.
				NewRule(
					Jump(ToUserDefinedChain(outboundRedirectChainName)),
				).
				WithComment("redirect all other outbound traffic to our custom chain for further processing"),
		)
	}

	return meshOutbound, nil
}

// buildMeshRedirect creates a chain in the NAT table to handle traffic redirection
// to a specified port. The chain will be configured to redirect TCP traffic to the
// provided port, which can be different for IPv4 and IPv6.
func buildMeshRedirect(cfg config.TrafficFlow, prefix string, ipv6 bool) (*Chain, error) {
	chainName := cfg.RedirectChain.GetFullName(prefix)

	// Determine the redirect port based on the IP version.
	redirectPort := cfg.Port
	if ipv6 && cfg.PortIPv6 != 0 {
		redirectPort = cfg.PortIPv6
	}

	// Create a new chain in the NAT table with the specified name.
	redirectChain, err := NewChain(TableNat, chainName)
	if err != nil {
		return nil, err
	}

	// Add a rule to redirect TCP traffic to the determined port.
	return redirectChain.AddRules(
		rules.
			NewRule(
				Protocol(Tcp()),
				Jump(ToPort(redirectPort)),
			).
			WithCommentf("redirect TCP traffic to envoy (port %d)", redirectPort),
	), nil
}

func addOutputRules(
	cfg config.InitializedConfig,
	dnsServers []string,
	nat *tables.NatTable,
	ipv6 bool,
) error {
	// Retrieve the fully qualified name of the outbound chain based on
	// configuration.
	outboundChainName := cfg.Redirect.Outbound.Chain.GetFullName(cfg.Redirect.NamePrefix)
	// DNS redirection port from configuration.
	dnsRedirectPort := cfg.Redirect.DNS.Port
	// Owner user ID for configuring UID-based rules.
	uid := cfg.Owner.UID
	// Initial position for the first rule in the NAT table.
	rulePosition := uint(1)

	shouldRedirectDNS := cfg.Redirect.DNS.EnabledIPv4
	if ipv6 {
		shouldRedirectDNS = cfg.Redirect.DNS.EnabledIPv6
	}

	// Add logging rule if logging is enabled in the configuration.
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
			// Return an error if the protocol is neither TCP nor UDP.
			return fmt.Errorf("unknown protocol %s, only 'tcp' or 'udp' allowed", uIDsToPorts.Protocol)
		}

		// Add rule to return early for specified ports and UID.
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
	if shouldRedirectDNS {
		// Determine if DOCKER_OUTPUT chain should be targeted based on IPv4/IPv6
		// and functionality.
		dockerOutputIPv6 := ipv6 && cfg.Executables.IPv6.Functionality.Chains.DockerOutput
		dockerOutputIPv4 := !ipv6 && cfg.Executables.IPv4.Functionality.Chains.DockerOutput
		dockerOutput := dockerOutputIPv6 || dockerOutputIPv4

		// Add DNS rule for redirecting DNS traffic based on UID.
		nat.Output().AddRules(
			rules.
				NewRule(
					Protocol(Udp(DestinationPort(DNSPort))),
					Match(Owner(Uid(uid))),
					JumpConditional(
						dockerOutput,                          // if DOCKER_OUTPUT should be targeted
						ToUserDefinedChain(ChainDockerOutput), // --jump DOCKER_OUTPUT
						Return(),                              // else RETURN
					),
				).
				WithPosition(rulePosition).
				WithConditionalComment(
					dockerOutput,
					fmt.Sprintf(
						"redirect DNS traffic from kuma-dp to the %s chain",
						ChainDockerOutput,
					),
					"return early for DNS traffic from kuma-dp",
				),
		)
		rulePosition++

		// Add rules to redirect all DNS requests or only those to specific
		// servers.
		if cfg.ShouldCaptureAllDNS() {
			nat.Output().AddRules(
				rules.
					NewRule(
						Protocol(Udp(DestinationPort(DNSPort))),
						Jump(ToPort(dnsRedirectPort)),
					).
					WithPosition(rulePosition).
					WithCommentf("redirect all DNS requests to the kuma-dp DNS proxy (listening on port %d)", dnsRedirectPort),
			)
		} else {
			for _, dnsIp := range dnsServers {
				nat.Output().AddRules(
					rules.
						NewRule(
							Destination(dnsIp),
							Protocol(Udp(DestinationPort(DNSPort))),
							Jump(ToPort(dnsRedirectPort)),
						).
						WithPosition(rulePosition).
						WithCommentf("redirect DNS requests to %s to the kuma-dp DNS proxy (listening on port %d)", dnsIp, dnsRedirectPort),
				)
				rulePosition++
			}
		}
	}

	// Add a default rule to direct all TCP traffic to the user-defined outbound
	// chain.
	nat.Output().AddRules(
		rules.
			NewRule(
				Protocol(Tcp()),
				Jump(ToUserDefinedChain(outboundChainName)),
			).
			WithComment("redirect outbound TCP traffic to our custom chain for processing"),
	)

	return nil
}

// addPreroutingRules adds rules to the PREROUTING chain of the NAT table to
// handle inbound traffic according to the provided configuration.
func addPreroutingRules(cfg config.InitializedConfig, nat *tables.NatTable, ipv6 bool) error {
	inboundChainName := cfg.Redirect.Inbound.Chain.GetFullName(cfg.Redirect.NamePrefix)
	rulePosition := uint(1)

	// Add a logging rule if logging is enabled.
	if cfg.Log.Enabled {
		nat.Prerouting().AddRules(
			rules.
				NewRule(Jump(Log(PreroutingLogPrefix, cfg.Log.Level))).
				WithComment("log matching packets using kernel logging"),
		)
	}

	// Handle virtual networks if they are defined in the configuration.
	if len(cfg.Redirect.VNet.Networks) > 0 {
		interfaceAndCidr := map[string]string{}
		for i := 0; i < len(cfg.Redirect.VNet.Networks); i++ {
			// Split the network definition into interface and CIDR.
			pair := strings.SplitN(cfg.Redirect.VNet.Networks[i], ":", 2)
			if len(pair) < 2 {
				return fmt.Errorf("incorrect definition of virtual network: %s", cfg.Redirect.VNet.Networks[i])
			}
			ipAddress, _, err := net.ParseCIDR(pair[1])
			if err != nil {
				return fmt.Errorf("incorrect CIDR definition: %s", err)
			}
			// Only include the address if it matches the IP version we are handling.
			if (ipv6 && ipAddress.To4() == nil) || (!ipv6 && ipAddress.To4() != nil) {
				interfaceAndCidr[pair[0]] = pair[1]
			}
		}
		for iface, cidr := range interfaceAndCidr {
			// Redirect DNS requests to the configured DNS port.
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
			)
			rulePosition++

			// Redirect all TCP traffic on specified virtual networks to the outbound port.
			nat.Prerouting().AddRules(
				rules.
					NewRule(
						NotDestination(cidr),
						InInterface(iface),
						Protocol(Tcp()),
						Jump(ToPort(cfg.Redirect.Outbound.Port)),
					).
					WithPosition(rulePosition).
					WithCommentf("redirect TCP traffic on interface %s, excluding destination %s, to the envoy's outbound passthrough port %d", iface, cidr, cfg.Redirect.Outbound.Port),
			)
			rulePosition++
		}
		nat.Prerouting().AddRules(
			// Redirect all remaining TCP traffic to the inbound chain for processing.
			rules.
				NewRule(
					Protocol(Tcp()),
					Jump(ToUserDefinedChain(inboundChainName)),
				).
				WithPosition(rulePosition).
				WithComment("redirect remaining TCP traffic to our custom chain for processing"),
		)
	} else {
		nat.Prerouting().AddRules(
			// Redirect inbound TCP traffic to the custom chain for processing.
			rules.
				NewRule(
					Protocol(Tcp()),
					Jump(ToUserDefinedChain(inboundChainName)),
				).
				WithComment("redirect inbound TCP traffic to our custom chain for processing"),
		)
	}
	return nil
}

// buildNatTable constructs the NAT table for iptables with the necessary rules
// for handling inbound and outbound traffic redirection, DNS redirection, and
// specific port exclusions or inclusions. It sets up custom chains for mesh
// traffic management based on the provided configuration.
func buildNatTable(cfg config.InitializedConfig, ipv6 bool) (*tables.NatTable, error) {
	prefix := cfg.Redirect.NamePrefix
	inboundRedirectChainName := cfg.Redirect.Inbound.RedirectChain.GetFullName(prefix)

	// Determine DNS servers based on IP version.
	dnsServers := cfg.Redirect.DNS.ServersIPv4
	if ipv6 {
		dnsServers = cfg.Redirect.DNS.ServersIPv6
	}

	// Initialize the NAT table.
	nat := tables.Nat()

	// Add output rules to the NAT table.
	if err := addOutputRules(cfg, dnsServers, nat, ipv6); err != nil {
		return nil, fmt.Errorf("could not add output rules %s", err)
	}

	// Add prerouting rules to the NAT table.
	if err := addPreroutingRules(cfg, nat, ipv6); err != nil {
		return nil, fmt.Errorf("could not add prerouting rules %s", err)
	}

	// Build the MESH_INBOUND chain.
	meshInbound, err := buildMeshInbound(cfg.Redirect.Inbound, prefix, inboundRedirectChainName)
	if err != nil {
		return nil, err
	}

	// Build the MESH_INBOUND_REDIRECT chain.
	meshInboundRedirect, err := buildMeshRedirect(cfg.Redirect.Inbound, prefix, ipv6)
	if err != nil {
		return nil, err
	}

	// Build the MESH_OUTBOUND chain.
	meshOutbound, err := buildMeshOutbound(cfg, dnsServers, ipv6)
	if err != nil {
		return nil, err
	}

	// Build the MESH_OUTBOUND_REDIRECT chain.
	meshOutboundRedirect, err := buildMeshRedirect(cfg.Redirect.Outbound, prefix, ipv6)
	if err != nil {
		return nil, err
	}

	// Add the custom chains to the NAT table and return it.
	return nat.
		WithCustomChain(meshInbound).
		WithCustomChain(meshOutbound).
		WithCustomChain(meshInboundRedirect).
		WithCustomChain(meshOutboundRedirect), nil
}
