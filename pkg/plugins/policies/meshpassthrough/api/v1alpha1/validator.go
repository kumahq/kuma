package v1alpha1

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"

	"github.com/asaskevich/govalidator"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

var (
	allMatchProtocols                = []string{string(TcpProtocol), string(TlsProtocol), string(GrpcProtocol), string(HttpProtocol), string(Http2Protocol)}
	notAllowedProtocolsOnTheSamePort = []ProtocolType{GrpcProtocol, HttpProtocol, Http2Protocol}
	wildcardPartialPrefixPattern     = regexp.MustCompile(`^\*[^\.]+`)
)

func (r *MeshPassthroughResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path.Field("default"), validateDefault(r.Spec.Default))
	return verr.OrNil()
}

func validateTop(targetRef *common_api.TargetRef) validators.ValidationError {
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
	portAndProtocol := map[uint32]ProtocolType{}
	type portProtocol struct {
		port     uint32
		protocol ProtocolType
	}
	uniqueDomains := map[portProtocol]map[string]bool{}
	for i, match := range conf.AppendMatch {
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
			if match.Protocol == "tcp" {
				verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("protocol"), "protocol tcp is not supported for a domain")
			}
			if wildcardPartialPrefixPattern.MatchString(match.Value) {
				verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("value"), "provided DNS has incorrect value, partial wildcard is currently not supported")
			}
			if match.Port == nil && strings.HasPrefix(match.Value, "*") && slices.Contains(notAllowedProtocolsOnTheSamePort, match.Protocol) {
				verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("port"), "wildcard domains doesn't work for all ports and layer 7 protocol")
			}
			if !strings.HasPrefix(match.Value, "*") {
				isValid := govalidator.IsDNSName(match.Value)
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
