package v1alpha1

import (
	"fmt"
	"math"
	"net"
	"regexp"
	"slices"
	"strings"

	"github.com/asaskevich/govalidator"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var (
	allMatchProtocols                = []string{string(TcpProtocol), string(TlsProtocol), string(GrpcProtocol), string(HttpProtocol), string(Http2Protocol), string(MysqlProtocol)}
	notAllowedProtocolsOnTheSamePort = []ProtocolType{GrpcProtocol, HttpProtocol, Http2Protocol}
	wildcardPartialPrefixPattern     = regexp.MustCompile(`^\*[^.]+`)
)

func (r *MeshPassthroughResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), r.validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path.Field("default"), validateDefault(r.Spec.Default))
	return verr.OrNil()
}

func (r *MeshPassthroughResource) validateTop(targetRef *common_api.TopLevelTargetRef) validators.ValidationError {
	if targetRef == nil {
		return validators.ValidationError{}
	}
	targetRefErr := mesh.ValidateTargetRef(targetRef.ToTargetRef(), &mesh.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.Dataplane,
		},
	})
	return targetRefErr
}

func validateDefault(conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	// matches accepted so far, a match without a port has port 0 and applies to all ports
	acceptedMatches := []chainMatch{}
	type portProtocol struct {
		port     uint32
		protocol ProtocolType
	}
	uniqueDomains := map[portProtocol]map[string]bool{}
	for i, match := range pointer.Deref(conf.AppendMatch) {
		if match.Protocol == MysqlProtocol && match.Port == nil {
			verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("port"), "port must be defined for Mysql protocol")
		}
		if match.Port != nil && pointer.Deref[uint32](match.Port) == 0 || pointer.Deref[uint32](match.Port) > math.MaxUint16 {
			verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("port"), "port must be a valid (1-65535)")
		}
		current := newChainMatch(match)
		if conflict, found := findChainConflict(acceptedMatches, current); found {
			field, message := conflictViolation(current, conflict)
			verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field(field), message)
		} else {
			acceptedMatches = append(acceptedMatches, current)
		}
		if match.Port != nil {
			key := portProtocol{
				port:     *match.Port,
				protocol: match.Protocol,
			}
			if _, found := uniqueDomains[key]; found {
				if _, found := uniqueDomains[key][match.Value]; found {
					verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("value"), fmt.Sprintf("value %s is already defiend for this port and protocol", match.Value))
				} else {
					uniqueDomains[key][match.Value] = true
				}
			} else {
				uniqueDomains[key] = map[string]bool{match.Value: true}
			}
		}
		if !slices.Contains(allMatchProtocols, string(match.Protocol)) {
			verr.AddErrorAt(validators.RootedAt("appendMatch").Index(i).Field("protocol"), validators.MakeFieldMustBeOneOfErr("protocol", allMatchProtocols...))
		}
		switch match.Type {
		case "CIDR":
			isValid := govalidator.IsCIDR(match.Value)
			if !isValid {
				verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("value"), "provided CIDR has incorrect value")
			}
		case "IP":
			isValid := govalidator.IsIP(match.Value)
			if !isValid {
				verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("value"), "provided IP has incorrect value")
			}
		case "Domain":
			if match.Protocol == "tcp" || match.Protocol == "mysql" {
				verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("protocol"), fmt.Sprintf("protocol %s is not supported for a domain", match.Protocol))
			}
			if wildcardPartialPrefixPattern.MatchString(match.Value) {
				verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("value"), "provided DNS has incorrect value, partial wildcard is currently not supported")
			}
			if match.Port == nil && strings.HasPrefix(match.Value, "*") && slices.Contains(notAllowedProtocolsOnTheSamePort, match.Protocol) {
				verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("port"), "wildcard domains doesn't work for all ports and layer 7 protocol")
			}
			valueToValidate := match.Value
			if strings.HasPrefix(match.Value, "*.") {
				valueToValidate = match.Value[2:]
			}
			if !strings.HasPrefix(valueToValidate, "*") {
				isValid := govalidator.IsDNSName(valueToValidate)
				if !isValid {
					verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("value"), "provided DNS has incorrect value")
				}
			}
		default:
			verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("type"), fmt.Sprintf("provided type %s is not supported, one of Domain, IP, or CIDR is supported", match.Type))
		}
	}
	return verr
}

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

// chainMatch identifies the filter chain a match resolves to: all domains of an L7
// protocol and port share one chain, TLS domains get a chain per SNI, IPs and CIDRs
// a chain per normalized destination prefix range.
type chainMatch struct {
	class    chainClass
	port     uint32
	address  string
	sni      string
	protocol ProtocolType
	value    string
}

func newChainMatch(match Match) chainMatch {
	result := chainMatch{
		class:    protocolClass(match.Protocol),
		port:     pointer.Deref[uint32](match.Port),
		protocol: match.Protocol,
		value:    match.Value,
	}
	switch match.Type {
	case "IP":
		if govalidator.IsIPv6(match.Value) {
			result.address = canonicalIP(match.Value) + "/128"
		} else {
			result.address = canonicalIP(match.Value) + "/32"
		}
	case "CIDR":
		result.address = canonicalCIDR(match.Value)
	case "Domain":
		if result.class == tlsChain {
			result.sni = match.Value
		}
	}
	return result
}

func canonicalIP(value string) string {
	if ip := net.ParseIP(value); ip != nil {
		return ip.String()
	}
	return value
}

func canonicalCIDR(value string) string {
	if _, ipNet, err := net.ParseCIDR(value); err == nil {
		return ipNet.String()
	}
	return value
}

// findChainConflict returns the first accepted match that resolves to the same filter
// chain as the given match
func findChainConflict(accepted []chainMatch, current chainMatch) (chainMatch, bool) {
	for _, match := range accepted {
		if match.class != current.class || match.address != current.address || match.sni != current.sni {
			continue
		}
		if match.port != current.port && match.port != 0 && current.port != 0 {
			continue
		}
		if current.address != "" && (match.protocol != current.protocol || match.value != current.value) {
			return match, true
		}
		// all domains of an L7 protocol and port share a single filter chain
		if current.address == "" && current.sni == "" && current.class == httpChain && match.protocol != current.protocol {
			return match, true
		}
	}
	return chainMatch{}, false
}

func conflictViolation(current chainMatch, conflict chainMatch) (string, string) {
	if current.protocol == conflict.protocol {
		return "value", fmt.Sprintf("match %q resolves to the same address as match %q on %s, both would produce the same filter chain", current.value, conflict.value, describeConflictPort(current, conflict))
	}
	if current.address != "" {
		return "protocol", fmt.Sprintf("protocols %s and %s for the same address on %s would produce the same filter chain, use a single protocol for %q", conflict.protocol, current.protocol, describeConflictPort(current, conflict), current.value)
	}
	if conflict.port == current.port {
		return "port", fmt.Sprintf("using the same port in multiple matches requires the same protocol for the following protocols: %v", notAllowedProtocolsOnTheSamePort)
	}
	// exactly one of the ports is undefined, a match without a port applies to all ports
	definedPort := current.port
	if definedPort == 0 {
		definedPort = conflict.port
	}
	return "port", fmt.Sprintf("a match without a port applies to all ports, so protocols %s and %s are both configured on port %d, using the same port in multiple matches requires the same protocol for the following protocols: %v", conflict.protocol, current.protocol, definedPort, notAllowedProtocolsOnTheSamePort)
}

func describeConflictPort(current chainMatch, conflict chainMatch) string {
	port := current.port
	if port == 0 {
		port = conflict.port
	}
	if port == 0 {
		return "all ports"
	}
	return fmt.Sprintf("port %d", port)
}
