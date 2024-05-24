package v1alpha1

import (
	"fmt"
	"math"

	"github.com/asaskevich/govalidator"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (r *MeshPassthroughResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path.Field("default"), validateDefault(r.Spec.Default))
	return verr.OrNil()
}

func validateTop(targetRef common_api.TargetRef) validators.ValidationError {
	targetRefErr := mesh.ValidateTargetRef(targetRef, &mesh.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.MeshSubset,
		},
	})
	return targetRefErr
}

func validateDefault(conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	for i, match := range conf.AppendMatch {
		if match.Port != nil && pointer.Deref[int](match.Port) == 0 || pointer.Deref[int](match.Port) > math.MaxUint16 {
			verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("port"), "port must be a valid (1-65535)")
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
			isValid := govalidator.IsDNSName(match.Value)
			if !isValid {
				verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("value"), "provided DNS has incorrect value")
			}
			if match.Protocol == "tcp" {
				verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("protocol"), "protocol tcp is not supported for a domain")
			}
		default:
			verr.AddViolationAt(validators.RootedAt("appendMatch").Index(i).Field("type"), fmt.Sprintf("provided type %s is not supported, one of Domain, IP, or CIDR is supported", match.Type))
		}
	}
	return verr
}
