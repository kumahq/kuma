package v1alpha1

import (
	"fmt"
	"maps"
	"net"
	"slices"
	"sort"
	"strings"

	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

// This file is the single source of truth for which matches of a MeshPassthrough
// configuration resolve to the same Envoy filter chain. The validator rejects such
// conflicts on apply, the generator drops them instead of producing a listener Envoy
// would reject, and MatchedPolicies surfaces the drops as warnings in the inspect API.

// chainClass groups protocols whose filter chains can collide: chains of different
// classes always differ in the transport or application protocols they match on.
type chainClass int

const (
	tcpChain  chainClass = iota // tcp and mysql: raw_buffer with optional address and port
	tlsChain                    // tls: tls transport with SNI or address and optional port
	httpChain                   // http, http2 and grpc: raw_buffer and http/1.1,h2c with optional address and port
)

func protocolClass(protocol ProtocolType) chainClass {
	switch protocol {
	case TlsProtocol:
		return tlsChain
	case HttpProtocol, Http2Protocol, GrpcProtocol:
		return httpChain
	default:
		return tcpChain
	}
}

// ChainKey identifies the filter chain a match resolves to: all domains of an L7
// protocol and port share one chain, TLS domains get a chain per SNI, IPs and CIDRs
// a chain per normalized destination prefix range.
type ChainKey struct {
	class   chainClass
	port    uint32
	address string
	sni     string
}

// KeyOf returns the identity of the filter chain the match resolves to.
func KeyOf(match Match) ChainKey {
	key := ChainKey{class: protocolClass(match.Protocol), port: pointer.Deref(match.Port)}
	switch match.Type {
	case "IP":
		if addressIsIPv6(match.Value) {
			key.address = canonicalIP(match.Value) + "/128"
		} else {
			key.address = canonicalIP(match.Value) + "/32"
		}
	case "CIDR":
		key.address = CanonicalCIDR(match.Value)
	case "Domain":
		if key.class == tlsChain {
			key.sni = match.Value
		}
	}
	return key
}

// WithPort returns the identity of the chain the same match produces on another port.
func (k ChainKey) WithPort(port uint32) ChainKey {
	k.port = port
	return k
}

// CanonicalCIDR normalizes a CIDR to the prefix range Envoy matches on: host bits
// dropped, non-canonical IPv6 spellings collapsed, IPv4-mapped IPv6 converted to
// plain IPv4. An unparseable value is returned as is, so distinct invalid values
// never resolve to one chain.
func CanonicalCIDR(value string) string {
	if _, ipNet, err := net.ParseCIDR(value); err == nil {
		return ipNet.String()
	}
	return value
}

func canonicalIP(value string) string {
	if ip := net.ParseIP(value); ip != nil {
		return ip.String()
	}
	return value
}

func addressIsIPv6(value string) bool {
	ip, _, _ := strings.Cut(value, "/")
	return govalidatorIsIPv6(ip)
}

// same rule the generator uses to pick the listener: textual IPv6 forms count as
// IPv6 even when they encode an IPv4 address
func govalidatorIsIPv6(ip string) bool {
	return net.ParseIP(ip) != nil && strings.Contains(ip, ":")
}

var l7Protocols = []ProtocolType{GrpcProtocol, HttpProtocol, Http2Protocol}

// chainValue is what distinguishes matches within one chain: L7 domains share the
// chain and merge as virtual host routes, everything else owns its chain exclusively.
func chainValue(match Match) string {
	if match.Type == "Domain" && slices.Contains(l7Protocols, match.Protocol) {
		return ""
	}
	return match.Value
}

type chainOwner struct {
	protocol ProtocolType
	value    string
	raw      string
}

type matchDrop struct {
	warning string
	// field and message anchor the validator violation; empty when other
	// validation already rejects the match
	field   string
	message string
}

type projection struct {
	key      ChainKey
	protocol ProtocolType
	value    string
}

// ConfAnalysis lists the matches of a Conf that cannot be applied: dropped matches
// resolve to a chain an earlier match already configures, suppressed projections are
// matches without a port not replicated onto a port whose chain a conflicting match
// owns. The first match configuring a chain always wins.
type ConfAnalysis struct {
	dropped    map[int]matchDrop
	suppressed map[projection]string
}

// AnalyzeConf mirrors how the generator assembles filter chains from AppendMatch.
func AnalyzeConf(conf Conf) ConfAnalysis {
	analysis := ConfAnalysis{dropped: map[int]matchDrop{}, suppressed: map[projection]string{}}
	owners := map[ChainKey]chainOwner{}
	ports := map[uint32]struct{}{}
	portless := map[projection]struct{}{}
	for i, match := range pointer.Deref(conf.AppendMatch) {
		if warning, invalid := invalidMatch(match); invalid {
			analysis.dropped[i] = matchDrop{warning: warning}
			continue
		}
		key := KeyOf(match)
		value := chainValue(match)
		owner, found := owners[key]
		if !found {
			owners[key] = chainOwner{protocol: match.Protocol, value: value, raw: match.Value}
			if port := pointer.Deref(match.Port); port == 0 {
				portless[projection{key: key, protocol: match.Protocol, value: value}] = struct{}{}
			} else {
				ports[port] = struct{}{}
			}
			continue
		}
		if owner.protocol == match.Protocol && owner.value == value {
			continue // the same chain, merges with the owner
		}
		analysis.dropped[i] = newMatchDrop(match, key, owner)
	}
	// a match without a port is replicated onto every port used elsewhere in the
	// policy, unless a conflicting match already owns the chain on that port
	for p := range portless {
		for port := range ports {
			key := p.key.WithPort(port)
			owner, found := owners[key]
			if !found || (owner.protocol == p.protocol && owner.value == p.value) {
				continue
			}
			var warning string
			if owner.protocol != p.protocol {
				warning = fmt.Sprintf(
					"matches with protocol %s and no port are not applied to %s on port %d, protocol %s is already configured there",
					p.protocol, key.describe(), port, owner.protocol,
				)
			} else {
				warning = fmt.Sprintf(
					"matches with protocol %s and no port are not applied to %s on port %d, match %q already defines a filter chain there",
					p.protocol, key.describe(), port, owner.raw,
				)
			}
			analysis.suppressed[projection{key: key, protocol: p.protocol, value: p.value}] = warning
		}
	}
	return analysis
}

// IsDropped reports whether the AppendMatch entry at the index resolves to a chain
// another match already configures and must not produce one of its own.
func (a ConfAnalysis) IsDropped(index int) bool {
	_, found := a.dropped[index]
	return found
}

// IsProjectionSuppressed reports whether a match without a port must not be
// replicated onto the chain identified by key.
func (a ConfAnalysis) IsProjectionSuppressed(key ChainKey, protocol ProtocolType, value string) bool {
	_, found := a.suppressed[projection{key: key, protocol: protocol, value: value}]
	return found
}

// Warnings renders one deduplicated, sorted message per dropped match and
// suppressed projection.
func (a ConfAnalysis) Warnings() []string {
	set := map[string]struct{}{}
	for _, drop := range a.dropped {
		set[drop.warning] = struct{}{}
	}
	for _, warning := range a.suppressed {
		set[warning] = struct{}{}
	}
	if len(set) == 0 {
		return nil
	}
	warnings := slices.Collect(maps.Keys(set))
	sort.Strings(warnings)
	return warnings
}

// ConfWarnings lists the warnings for every match the generator drops from the
// configuration. The policy matcher attaches them to the matched policies, which
// is what the dataplane `_rules` inspect API serves.
func (c Conf) ConfWarnings() []string {
	return AnalyzeConf(c).Warnings()
}

// invalidMatch drops matches the generator cannot turn into a valid filter chain.
// The validator rejects them on apply, so they can only appear when validation was
// bypassed, e.g. in a policy stored before an upgrade.
func invalidMatch(match Match) (string, bool) {
	switch match.Type {
	case "IP":
		if net.ParseIP(match.Value) == nil {
			return fmt.Sprintf("ignoring match %q, not a valid IP", match.Value), true
		}
	case "CIDR":
		if _, _, err := net.ParseCIDR(match.Value); err != nil {
			return fmt.Sprintf("ignoring match %q, not a valid CIDR", match.Value), true
		}
	case "Domain":
	default:
		return fmt.Sprintf("ignoring match %q, type %q is not supported", match.Value, match.Type), true
	}
	return "", false
}

func newMatchDrop(match Match, key ChainKey, owner chainOwner) matchDrop {
	port := pointer.Deref(match.Port)
	if owner.protocol != match.Protocol {
		drop := matchDrop{
			warning: fmt.Sprintf(
				"ignoring match %q with protocol %s, protocol %s is already configured for %s on %s, %s",
				match.Value, match.Protocol, owner.protocol, key.describe(), describePort(port), key.conflictReason(),
			),
			field: "protocol",
		}
		switch {
		case key.address != "":
			drop.message = fmt.Sprintf(
				"protocols %s and %s for the same address on %s would produce the same filter chain, use a single protocol for %q",
				owner.protocol, match.Protocol, describePort(port), match.Value,
			)
		case key.class == httpChain && port != 0:
			drop.field = "port"
			drop.message = fmt.Sprintf("using the same port in multiple matches requires the same protocol for the following protocols: %v", l7Protocols)
		case key.class == httpChain:
			drop.message = fmt.Sprintf("protocols %s and %s share a single filter chain for domains without a port, use a single protocol", owner.protocol, match.Protocol)
		default:
			drop.message = fmt.Sprintf("protocols %s and %s on %s would produce the same filter chain for %q", owner.protocol, match.Protocol, describePort(port), match.Value)
		}
		return drop
	}
	return matchDrop{
		warning: fmt.Sprintf(
			"ignoring match %q with protocol %s, match %q already defines a filter chain for %s on %s",
			match.Value, match.Protocol, owner.raw, key.describe(), describePort(port),
		),
		field: "value",
		message: fmt.Sprintf(
			"match %q resolves to the same address as match %q on %s, both would produce the same filter chain",
			match.Value, owner.raw, describePort(port),
		),
	}
}

func (k ChainKey) conflictReason() string {
	if k.class == httpChain {
		return fmt.Sprintf("only one of %v can be configured on the same port", l7Protocols)
	}
	return "both would produce the same filter chain matcher"
}

func (k ChainKey) describe() string {
	if k.sni != "" {
		return fmt.Sprintf("domain %q", k.sni)
	}
	if k.address == "" {
		return "domains"
	}
	return k.address
}

func describePort(port uint32) string {
	if port == 0 {
		return "all ports"
	}
	return fmt.Sprintf("port %d", port)
}
