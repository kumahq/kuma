package v1alpha1

import (
	"fmt"
	"net"
	"slices"

	"github.com/asaskevich/govalidator"

	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var l7Protocols = []ProtocolType{GrpcProtocol, HttpProtocol, Http2Protocol}

// chains of the same class are only told apart by port and address, chains of
// different classes always differ in the transport or application protocols they
// match on: raw_buffer for tcp and mysql, the tls transport for tls, raw_buffer with
// http/1.1 and h2c for http, http2 and grpc
type chainClass int

const (
	tcpChain chainClass = iota
	tlsChain
	httpChain
)

func protocolClass(protocol ProtocolType) chainClass {
	switch {
	case protocol == TlsProtocol:
		return tlsChain
	case slices.Contains(l7Protocols, protocol):
		return httpChain
	default:
		return tcpChain
	}
}

// FilterChainMatcher identifies the filter chain a match resolves to: all domains
// of an L7 protocol and port share one chain, TLS domains get a chain per SNI, IPs
// and CIDRs a chain per normalized address range.
type FilterChainMatcher struct {
	class   chainClass
	port    uint32
	address string
	sni     string
}

func (m Match) FilterChainMatcher() FilterChainMatcher {
	matcher := FilterChainMatcher{class: protocolClass(m.Protocol), port: pointer.Deref(m.Port)}
	switch m.Type {
	case "IP":
		// the generator picks the listener the same way, a textual IPv6 form is IPv6
		// even when it encodes an IPv4 address
		if govalidator.IsIPv6(m.Value) {
			matcher.address = CanonicalCIDR(m.Value + "/128")
		} else {
			matcher.address = CanonicalCIDR(m.Value + "/32")
		}
	case "CIDR":
		matcher.address = CanonicalCIDR(m.Value)
	case "Domain":
		if matcher.class == tlsChain {
			matcher.sni = m.Value
		}
	}
	return matcher
}

func (f FilterChainMatcher) WithPort(port uint32) FilterChainMatcher {
	f.port = port
	return f
}

func (f FilterChainMatcher) String() string {
	chain := "domains"
	switch {
	case f.sni != "":
		chain = fmt.Sprintf("domain %q", f.sni)
	case f.address != "":
		chain = f.address
	}
	if f.port == 0 {
		return chain + " on all ports"
	}
	return fmt.Sprintf("%s on port %d", chain, f.port)
}

// CanonicalCIDR returns the prefix range Envoy matches on, host bits dropped and
// IPv4-mapped IPv6 collapsed to plain IPv4. An unparseable value is returned as is,
// so two invalid values never resolve to one chain.
func CanonicalCIDR(value string) string {
	if _, ipNet, err := net.ParseCIDR(value); err == nil {
		return ipNet.String()
	}
	return value
}

// the match a filter chain is built from
type chainOwner struct {
	protocol ProtocolType
	// the part of the chain the match owns, empty for L7 domains: they share their
	// chain and merge as virtual host routes, anything else owns it exclusively
	chainValue string
	value      string // the match value as written, for messages
}

func chainOwnerOf(match Match) chainOwner {
	owner := chainOwner{protocol: match.Protocol, chainValue: match.Value, value: match.Value}
	if match.Type == "Domain" && slices.Contains(l7Protocols, match.Protocol) {
		owner.chainValue = ""
	}
	return owner
}

func (o chainOwner) sharesChainWith(other chainOwner) bool {
	return o.protocol == other.protocol && o.chainValue == other.chainValue
}

// field and message are empty when the match is invalid rather than conflicting,
// other validation already rejects those
type conflict struct {
	field   string
	message string
}

// Conflicts lists what a Conf cannot apply: dropped matches resolve to a chain an
// earlier match already configures, suppressed chains are the ones a match without a
// port is not replicated onto. The validator rejects both on apply, the generator
// drops them instead of building a listener Envoy rejects.
type Conflicts struct {
	dropped    map[int]conflict
	suppressed map[FilterChainMatcher]struct{}
	warnings   []string
}

// FindConflicts mirrors how the generator assembles filter chains from AppendMatch.
// The first match configuring a chain wins.
func FindConflicts(conf Conf) Conflicts {
	conflicts := Conflicts{dropped: map[int]conflict{}, suppressed: map[FilterChainMatcher]struct{}{}}
	owners := map[FilterChainMatcher]chainOwner{}
	var ports []uint32
	var portless []FilterChainMatcher
	for i, match := range pointer.Deref(conf.AppendMatch) {
		if reason, invalid := invalidMatch(match); invalid {
			conflicts.dropped[i] = conflict{}
			conflicts.warn("ignoring match %q, %s", match.Value, reason)
			continue
		}
		chain := match.FilterChainMatcher()
		candidate := chainOwnerOf(match)
		owner, used := owners[chain]
		switch {
		case !used:
			owners[chain] = candidate
			switch {
			case chain.port == 0:
				portless = append(portless, chain)
			case !slices.Contains(ports, chain.port):
				ports = append(ports, chain.port)
			}
		case owner.sharesChainWith(candidate):
			// the same chain, merges with the owner
		default:
			field := "value"
			if owner.protocol != candidate.protocol {
				field = "protocol"
			}
			reason := conflictReason(owner, candidate, chain)
			conflicts.dropped[i] = conflict{field: field, message: reason}
			conflicts.warn("ignoring match %q, %s", match.Value, reason)
		}
	}
	conflicts.dropPortlessCopies(owners, portless, ports)
	return conflicts
}

// a match without a port is copied onto every port used elsewhere in the policy,
// except ports whose chain a conflicting match already owns
func (c *Conflicts) dropPortlessCopies(owners map[FilterChainMatcher]chainOwner, portless []FilterChainMatcher, ports []uint32) {
	for _, chain := range portless {
		copied := owners[chain]
		for _, port := range ports {
			target := chain.WithPort(port)
			owner, used := owners[target]
			if !used || owner.sharesChainWith(copied) {
				continue
			}
			c.suppressed[target] = struct{}{}
			c.warn(
				"%s, matches with protocol %s and no port are not applied there",
				conflictReason(owner, copied, target), copied.protocol,
			)
		}
	}
}

func conflictReason(owner, candidate chainOwner, chain FilterChainMatcher) string {
	if owner.protocol != candidate.protocol {
		return fmt.Sprintf("protocols %s and %s produce the same filter chain for %s", owner.protocol, candidate.protocol, chain)
	}
	return fmt.Sprintf("matches %q and %q produce the same filter chain for %s", owner.value, candidate.value, chain)
}

func (c *Conflicts) warn(format string, args ...any) {
	warning := fmt.Sprintf(format, args...)
	if !slices.Contains(c.warnings, warning) {
		c.warnings = append(c.warnings, warning)
	}
}

func (c Conflicts) IsDropped(index int) bool {
	_, found := c.dropped[index]
	return found
}

func (c Conflicts) IsSuppressed(matcher FilterChainMatcher) bool {
	_, found := c.suppressed[matcher]
	return found
}

func (c Conflicts) Warnings() []string {
	return c.warnings
}

func (c Conf) Warnings() []string {
	return FindConflicts(c).Warnings()
}

// invalidMatch reports matches the generator cannot turn into a filter chain, they
// only show up when validation was bypassed, for example in a policy stored before an
// upgrade.
func invalidMatch(match Match) (string, bool) {
	switch match.Type {
	case "IP":
		if net.ParseIP(match.Value) == nil {
			return "not a valid IP", true
		}
	case "CIDR":
		if _, _, err := net.ParseCIDR(match.Value); err != nil {
			return "not a valid CIDR", true
		}
	case "Domain":
	default:
		return fmt.Sprintf("type %q is not supported", match.Type), true
	}
	return "", false
}
