package v1alpha1

import (
	"fmt"
	"math"
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
	// L7 matches accepted so far, a match without a port has port 0 and applies to all ports
	acceptedL7Matches := []l7Match{}
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
		if slices.Contains(notAllowedProtocolsOnTheSamePort, match.Protocol) {
			current := newL7Match(match)
			if conflict, found := findL7Conflict(acceptedL7Matches, current); found {
				verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("port"), conflictMessage(current, conflict))
			} else {
				acceptedL7Matches = append(acceptedL7Matches, current)
			}
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

// l7Match is a match with a protocol that cannot share a filter chain with another L7
// protocol. Port 0 means the match has no port defined and applies to all ports. All
// domains of a given protocol and port share a single filter chain, so they conflict with
// each other, IPs and CIDRs are matched on the destination address and conflict only with
// the same address.
type l7Match struct {
	port     uint32
	address  string
	protocol ProtocolType
}

func newL7Match(match Match) l7Match {
	result := l7Match{
		port:     pointer.Deref[uint32](match.Port),
		protocol: match.Protocol,
	}
	if match.Type != "Domain" {
		result.address = match.Value
	}
	return result
}

// findL7Conflict returns the first accepted match that ends up in the same filter chain
// as the given match but with a different protocol
func findL7Conflict(accepted []l7Match, current l7Match) (l7Match, bool) {
	for _, match := range accepted {
		if match.protocol == current.protocol || match.address != current.address {
			continue
		}
		if match.port == current.port || match.port == 0 || current.port == 0 {
			return match, true
		}
	}
	return l7Match{}, false
}

func conflictMessage(current l7Match, conflict l7Match) string {
	if conflict.port == current.port {
		return fmt.Sprintf("using the same port in multiple matches requires the same protocol for the following protocols: %v", notAllowedProtocolsOnTheSamePort)
	}
	// exactly one of the ports is undefined, a match without a port applies to all ports
	definedPort := current.port
	if definedPort == 0 {
		definedPort = conflict.port
	}
	return fmt.Sprintf("a match without a port applies to all ports, so protocols %s and %s are both configured on port %d, using the same port in multiple matches requires the same protocol for the following protocols: %v", conflict.protocol, current.protocol, definedPort, notAllowedProtocolsOnTheSamePort)
}
