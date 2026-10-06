package v1alpha1

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"

	"github.com/asaskevich/govalidator"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/validators"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
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

func (r *MeshPassthroughResource) validateTop(targetRef *common_api.TargetRef) validators.ValidationError {
	if targetRef == nil {
		return validators.ValidationError{}
	}
	targetRefErr := mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.MeshSubset,
			common_api.Dataplane,
		},
	})
	return targetRefErr
}

func validateDefault(conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	// http, http2 and grpc build the same filter chain match, other protocols differ
	// in the transport or application protocol, so only L7 protocols are compared
	l7ProtocolOnPort := map[uint32]ProtocolType{}
	type portProtocol struct {
		port     uint32
		protocol ProtocolType
	}
	uniqueDomains := map[portProtocol]map[string]bool{}
	type chainMatch struct {
		index    int
		protocol ProtocolType
		key      string
	}
	var portlessMatches []chainMatch
	var ports []uint32
	matchesOnPort := map[uint32][]chainMatch{}
	for i, match := range pointer.Deref(conf.AppendMatch) {
		if match.Protocol == MysqlProtocol && match.Port == nil {
			verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("port"), "port must be defined for Mysql protocol")
		}
		if match.Port != nil && pointer.Deref[uint32](match.Port) == 0 || pointer.Deref[uint32](match.Port) > math.MaxUint16 {
			verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("port"), "port must be a valid (1-65535)")
		}
		// matches without a port share port 0, the generator rejects different L7 protocols there too
		if slices.Contains(notAllowedProtocolsOnTheSamePort, match.Protocol) {
			if protocol, found := l7ProtocolOnPort[pointer.Deref(match.Port)]; !found {
				l7ProtocolOnPort[pointer.Deref(match.Port)] = match.Protocol
			} else if protocol != match.Protocol {
				verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("port"), fmt.Sprintf("using the same port in multiple matches requires the same protocol for the following protocols: %v", notAllowedProtocolsOnTheSamePort))
			}
		}
		chain := chainMatch{index: i, protocol: match.Protocol, key: filterChainKey(match)}
		if match.Port == nil {
			portlessMatches = append(portlessMatches, chain)
		} else {
			if _, found := matchesOnPort[*match.Port]; !found {
				ports = append(ports, *match.Port)
			}
			matchesOnPort[*match.Port] = append(matchesOnPort[*match.Port], chain)
		}
		if match.Port != nil {
			key := portProtocol{
				port:     *match.Port,
				protocol: match.Protocol,
			}
			if uniqueDomains[key][match.Value] {
				verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("value"), fmt.Sprintf("value %s is already defined for this port and protocol", match.Value))
			} else {
				keys := []portProtocol{key}
				// tcp and mysql build the same filter chain match, so the value is taken for both
				if match.Protocol == TcpProtocol || match.Protocol == MysqlProtocol {
					keys = []portProtocol{{port: key.port, protocol: TcpProtocol}, {port: key.port, protocol: MysqlProtocol}}
				}
				for _, key := range keys {
					if uniqueDomains[key] == nil {
						uniqueDomains[key] = map[string]bool{}
					}
					uniqueDomains[key][match.Value] = true
				}
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
	// the generator copies a match without a port onto every port used by other matches
	for _, portless := range portlessMatches {
		for _, port := range ports {
			idx := slices.IndexFunc(matchesOnPort[port], func(other chainMatch) bool {
				return other.key == portless.key && other.protocol != portless.protocol
			})
			if idx >= 0 {
				other := matchesOnPort[port][idx]
				verr.AddViolationAt(validators.RootedAt("appendMatch").Index(portless.index).Field("port"), fmt.Sprintf("a match without a port is also applied to port %d, where appendMatch[%d] with protocol %s builds the same filter chain", port, other.index, other.protocol))
				break
			}
		}
	}
	return verr
}

// filterChainKey identifies the filter chain match a match builds, apart from the port:
// L7 domains share one chain per protocol, other matches get a chain per value
func filterChainKey(match Match) string {
	l7 := slices.Contains(notAllowedProtocolsOnTheSamePort, match.Protocol)
	switch {
	case l7 && match.Type == "Domain":
		return "l7"
	case l7:
		return "l7/" + match.Value
	case match.Protocol == MysqlProtocol:
		return string(TcpProtocol) + "/" + match.Value
	default:
		return string(match.Protocol) + "/" + match.Value
	}
}
