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

// ChainKey identifies the filter chain a match resolves to: all domains of an L7
// protocol and port share one chain, TLS domains get a chain per SNI, IPs and CIDRs
// a chain per normalized address range.
type ChainKey struct {
	class   chainClass
	port    uint32
	address string
	sni     string
}

func (m Match) ChainKey() ChainKey {
	key := ChainKey{class: protocolClass(m.Protocol), port: pointer.Deref(m.Port)}
	switch m.Type {
	case "IP":
		// the generator picks the listener the same way, a textual IPv6 form is IPv6
		// even when it encodes an IPv4 address
		if govalidator.IsIPv6(m.Value) {
			key.address = CanonicalCIDR(m.Value + "/128")
		} else {
			key.address = CanonicalCIDR(m.Value + "/32")
		}
	case "CIDR":
		key.address = CanonicalCIDR(m.Value)
	case "Domain":
		if key.class == tlsChain {
			key.sni = m.Value
		}
	}
	return key
}

func (k ChainKey) WithPort(port uint32) ChainKey {
	k.port = port
	return k
}

func (k ChainKey) String() string {
	chain := "domains"
	switch {
	case k.sni != "":
		chain = fmt.Sprintf("domain %q", k.sni)
	case k.address != "":
		chain = k.address
	}
	if k.port == 0 {
		return chain + " on all ports"
	}
	return fmt.Sprintf("%s on port %d", chain, k.port)
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

// L7 domains share their chain and merge as virtual host routes, anything else owns
// its chain exclusively
func chainValue(match Match) string {
	if match.Type == "Domain" && slices.Contains(l7Protocols, match.Protocol) {
		return ""
	}
	return match.Value
}

type chainOwner struct {
	protocol ProtocolType
	value    string
	raw      string // the value as written, for messages
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
	suppressed map[ChainKey]struct{}
	warnings   []string
}

// FindConflicts mirrors how the generator assembles filter chains from AppendMatch.
// The first match configuring a chain wins.
func FindConflicts(conf Conf) Conflicts {
	conflicts := Conflicts{dropped: map[int]conflict{}, suppressed: map[ChainKey]struct{}{}}
	owners := map[ChainKey]chainOwner{}
	var ports []uint32
	var portless []ChainKey
	for i, match := range pointer.Deref(conf.AppendMatch) {
		if reason, invalid := invalidMatch(match); invalid {
			conflicts.dropped[i] = conflict{}
			conflicts.warn("ignoring match %q, %s", match.Value, reason)
			continue
		}
		key := match.ChainKey()
		value := chainValue(match)
		owner, found := owners[key]
		if !found {
			owners[key] = chainOwner{protocol: match.Protocol, value: value, raw: match.Value}
			switch {
			case key.port == 0:
				portless = append(portless, key)
			case !slices.Contains(ports, key.port):
				ports = append(ports, key.port)
			}
			continue
		}
		if owner.protocol == match.Protocol && owner.value == value {
			continue // the same chain, merges with the owner
		}
		field := "value"
		if owner.protocol != match.Protocol {
			field = "protocol"
		}
		reason := conflictReason(owner, match.Protocol, match.Value, key)
		conflicts.dropped[i] = conflict{field: field, message: reason}
		conflicts.warn("ignoring match %q, %s", match.Value, reason)
	}
	// a match without a port is replicated onto every port used elsewhere in the
	// policy, except ports whose chain a conflicting match already owns
	for _, key := range portless {
		projected := owners[key]
		for _, port := range ports {
			target := key.WithPort(port)
			owner, found := owners[target]
			if !found || (owner.protocol == projected.protocol && owner.value == projected.value) {
				continue
			}
			conflicts.suppressed[target] = struct{}{}
			conflicts.warn(
				"%s, matches with protocol %s and no port are not applied there",
				conflictReason(owner, projected.protocol, projected.raw, target), projected.protocol,
			)
		}
	}
	return conflicts
}

func conflictReason(owner chainOwner, protocol ProtocolType, value string, key ChainKey) string {
	if owner.protocol != protocol {
		return fmt.Sprintf("protocols %s and %s produce the same filter chain for %s", owner.protocol, protocol, key)
	}
	return fmt.Sprintf("matches %q and %q produce the same filter chain for %s", owner.raw, value, key)
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

func (c Conflicts) IsSuppressed(key ChainKey) bool {
	_, found := c.suppressed[key]
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
