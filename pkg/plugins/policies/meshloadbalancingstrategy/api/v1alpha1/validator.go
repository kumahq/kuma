package v1alpha1

import (
	"strings"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *MeshLoadBalancingStrategyResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	if len(r.Spec.To) == 0 {
		verr.AddViolationAt(path.Field("to"), "needs at least one item")
	}
	verr.AddErrorAt(path, validateTo(r.Spec.To))
	return verr.OrNil()
}

func validateTop(targetRef common_api.TargetRef) validators.ValidationError {
	targetRefErr := mesh.ValidateTargetRef(targetRef, &mesh.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.MeshSubset,
			common_api.MeshService,
			common_api.MeshServiceSubset,
		},
	})
	return targetRefErr
}

func validateTo(to []To) validators.ValidationError {
	var verr validators.ValidationError
	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), mesh.ValidateTargetRef(toItem.TargetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshService,
			},
		}))
		verr.AddErrorAt(path.Field("default"), validateDefault(toItem.Default))
	}
	return verr
}

func validateDefault(conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	verr.AddError("loadBalancer", validateLoadBalancer(conf.LoadBalancer))
	return verr
}

func validateLoadBalancer(conf *LoadBalancer) validators.ValidationError {
	var verr validators.ValidationError
	if conf == nil {
		return verr
	}

	switch conf.Type {
	case RingHashType:
		if conf.RingHash != nil {
			verr.AddError("ringHash", validateRingHash(conf.RingHash))
		}
	case MaglevType:
		if conf.Maglev != nil {
			verr.AddError("maglev", validateMaglev(conf.Maglev))
		}
	case RoundRobinType:
	case RandomType:
	case LeastRequestType:
	}

	return verr
}

func validateRingHash(conf *RingHash) validators.ValidationError {
	var verr validators.ValidationError
	if conf == nil {
		return verr
	}
	verr.AddError("hashPolicies", validateHashPolicies(conf.HashPolicies))
	return verr
}

func validateMaglev(conf *Maglev) validators.ValidationError {
	var verr validators.ValidationError
	if conf == nil {
		return verr
	}
	verr.AddError("hashPolicies", validateHashPolicies(conf.HashPolicies))
	return verr
}

func validateHashPolicies(conf *[]HashPolicy) validators.ValidationError {
	var verr validators.ValidationError
	if conf == nil {
		return verr
	}
	for idx, policy := range *conf {
		path := validators.Root().Index(idx)
		switch policy.Type {
		case HeaderType:
			if policy.Header == nil {
				verr.AddViolationAt(path.Field("header"), validators.MustBeDefined)
			}
		case CookieType:
			if policy.Cookie == nil {
				verr.AddViolationAt(path.Field("cookie"), validators.MustBeDefined)
			} else if policy.Cookie.Path != nil && !strings.HasPrefix(*policy.Cookie.Path, "/") {
				verr.AddViolationAt(path.Field("cookie").Field("path"), "must be an absolute path")
			}
		case QueryParameterType:
			if policy.QueryParameter == nil {
				verr.AddViolationAt(path.Field("queryParameter"), validators.MustBeDefined)
			}
		case FilterStateType:
			if policy.FilterState == nil {
				verr.AddViolationAt(path.Field("filterState"), validators.MustBeDefined)
			}
		case ConnectionType:
			if policy.Connection == nil {
				verr.AddViolationAt(path.Field("connection"), validators.MustBeDefined)
			}
		}
	}
	return verr
}
