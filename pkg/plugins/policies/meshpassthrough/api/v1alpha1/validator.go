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

func (r *MeshPassthroughResource) validateTop(targetRef *common_api.TargetRef) validators.ValidationError {
	if targetRef == nil {
		return validators.ValidationError{}
	}
	targetRefErr := mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.Dataplane,
		},
	})
	return targetRefErr
}

func validateDefault(conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	portAndProtocol := map[uint32]ProtocolType{}
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
		if match.Port != nil {
			if value, found := portAndProtocol[pointer.Deref[uint32](match.Port)]; found && value != match.Protocol && slices.Contains(notAllowedProtocolsOnTheSamePort, match.Protocol) {
				verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("port"), fmt.Sprintf("using the same port in multiple matches requires the same protocol for the following protocols: %v", notAllowedProtocolsOnTheSamePort))
			} else {
				portAndProtocol[pointer.Deref[uint32](match.Port)] = match.Protocol
			}
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
	verr.Add(validateAllPortsL7Conflicts(pointer.Deref(conf.AppendMatch)))
	return verr
}

// validateAllPortsL7Conflicts rejects configurations where an L7 match (http/http2/grpc)
// without a port collides with a different L7 protocol on the same "scope". http, http2 and
// grpc share identical filter-chain matching (transport raw_buffer + application protocols
// http/1.1,h2c), so two of them cannot select a filter chain on the same port. A match with no
// port applies to every port, so it conflicts with any different-protocol L7 match regardless of
// that match's port. This case is missed by the per-port check above (which only inspects
// matches that declare a port) and, left unvalidated, produces a listener Envoy rejects with a
// duplicate filter-chain match error (e.g. the same domain configured for both grpc and http).
//
// Matches are grouped by scope: all Domain matches share one scope (their filter chains are named
// by protocol+port only), while each IP/CIDR value is its own scope (its chains embed the
// address). Only a shared scope can collide.
func validateAllPortsL7Conflicts(matches []Match) validators.ValidationError {
	var verr validators.ValidationError

	scopeOf := func(m Match) (string, bool) {
		if !slices.Contains(notAllowedProtocolsOnTheSamePort, m.Protocol) {
			return "", false
		}
		switch m.Type {
		case "Domain":
			return "domain", true
		case "IP", "CIDR":
			return string(m.Type) + ":" + m.Value, true
		default:
			return "", false
		}
	}

	scopeProtocols := map[string]map[ProtocolType]bool{}
	for _, match := range matches {
		if scope, ok := scopeOf(match); ok {
			if scopeProtocols[scope] == nil {
				scopeProtocols[scope] = map[ProtocolType]bool{}
			}
			scopeProtocols[scope][match.Protocol] = true
		}
	}

	for i, match := range matches {
		scope, ok := scopeOf(match)
		if !ok || match.Port != nil {
			continue
		}
		var conflicting []string
		for protocol := range scopeProtocols[scope] {
			if protocol != match.Protocol {
				conflicting = append(conflicting, string(protocol))
			}
		}
		if len(conflicting) > 0 {
			slices.Sort(conflicting)
			verr.AddViolationAt(
				validators.RootedAt("appendMatch").Index(i).Field("port"),
				fmt.Sprintf("protocol %s without a port cannot be combined with protocol %s on the same destination; %v cannot share a port", match.Protocol, conflicting[0], notAllowedProtocolsOnTheSamePort),
			)
		}
	}
	return verr
}
