package v1alpha1

import (
	"fmt"
	"strings"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (r *MeshLoadBalancingStrategyResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	if len(r.Spec.To) == 0 {
		verr.AddViolationAt(path.Field("to"), "needs at least one item")
	}
	topLevel := pointer.DerefOr(r.Spec.TargetRef, common_api.TargetRef{Kind: common_api.Mesh})
	verr.AddErrorAt(path, validateTo(topLevel, r.Spec.To))
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
			common_api.MeshGateway,
			common_api.MeshService,
			common_api.MeshServiceSubset,
		},
		GatewayListenerTagsAllowed: true,
	})
	return targetRefErr
}

func validateTo(topTargetRef common_api.TargetRef, to []To) validators.ValidationError {
	var verr validators.ValidationError
	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)
		var supportedKinds []common_api.TargetRefKind
		var supportedKindsError string
		switch topTargetRef.Kind {
		case common_api.MeshGateway:
			if toItem.Default.LoadBalancer != nil {
				supportedKindsError = fmt.Sprintf("value is not supported, only %s is allowed if loadBalancer is set", common_api.Mesh)
				supportedKinds = []common_api.TargetRefKind{
					common_api.Mesh,
				}
			} else {
				supportedKinds = []common_api.TargetRefKind{
					common_api.Mesh,
					common_api.MeshService,
					common_api.MeshMultiZoneService,
				}
			}
		default:
			supportedKinds = []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshService,
				common_api.MeshMultiZoneService,
			}
		}
		errs := mesh.ValidateTargetRef(toItem.TargetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds:      supportedKinds,
			SupportedKindsError: supportedKindsError,
		})
		verr.AddErrorAt(path.Field("targetRef"), errs)
		if toItem.TargetRef.Kind == common_api.MeshExternalService && topTargetRef.Kind != common_api.Mesh {
			verr.AddViolationAt(path.Field("targetRef.kind"), "kind MeshExternalService is only allowed with targetRef.kind: Mesh as it is configured on the Zone Egress and shared by all clients in the mesh")
		}
		verr.AddErrorAt(path.Field("default"), validateConf(toItem.Default, toItem))
	}
	return verr
}

func validateConf(conf Conf, to To) validators.ValidationError {
	var verr validators.ValidationError
	verr.AddError("loadBalancer", validateLoadBalancer(conf.LoadBalancer))
	verr.AddError("localityAwareness", validateLocalityAwareness(conf.LocalityAwareness, to))
	return verr
}

func validateLocalityAwareness(localityAwareness *LocalityAwareness, to To) validators.ValidationError {
	var verr validators.ValidationError
	if localityAwareness == nil {
		return verr
	}
	verr.AddError("localZone", validateLocalZone(localityAwareness.LocalZone))
	verr.AddError("crossZone", validateCrossZone(localityAwareness.CrossZone, to))
	return verr
}

func validateLocalZone(localZone *LocalZone) validators.ValidationError {
	var verr validators.ValidationError
	if localZone == nil {
		return verr
	}

	var weightSpecified int
	for idx, affinityTag := range pointer.Deref(localZone.AffinityTags) {
		path := validators.RootedAt("affinityTags").Index(idx)
		if affinityTag.Key == "" {
			verr.AddViolationAt(path.Field("key"), validators.MustNotBeEmpty)
		}
		if affinityTag.Weight != nil {
			verr.Add(validators.ValidateIntegerGreaterThanZeroOrNil(path.Field("weight"), affinityTag.Weight))
			weightSpecified++
		}
	}

	if weightSpecified > 0 && weightSpecified != len(pointer.Deref(localZone.AffinityTags)) {
		verr.AddViolation("affinityTags", "all or none affinity tags should have weight")
	}
	return verr
}

func validateCrossZone(crossZone *CrossZone, to To) validators.ValidationError {
	var verr validators.ValidationError
	if crossZone == nil {
		return verr
	}
	if to.TargetRef.Kind == common_api.MeshService && (to.TargetRef.SectionName != "" || len(to.TargetRef.Labels) > 0) {
		verr.AddViolationAt(validators.Root(), fmt.Sprintf("%s: MeshService traffic is local", validators.MustNotBeSet))
	}

	for idx, failover := range crossZone.Failover {
		path := validators.RootedAt("failover").Index(idx)
		if failover.From != nil {
			if len(failover.From.Zones) == 0 {
				verr.AddViolationAt(path.Field("from").Field("zones"), validators.MustNotBeEmpty)
			}

			for zoneIdx, from := range failover.From.Zones {
				if from == "" {
					verr.AddViolationAt(path.Field("from").Field("zones").Index(zoneIdx), validators.MustNotBeEmpty)
				}
			}
		}

		toZonesPath := path.Field("to").Field("zones")
		switch failover.To.Type {
		case Any, None:
			if failover.To.Zones != nil && len(*failover.To.Zones) > 0 {
				verr.AddViolationAt(toZonesPath, fmt.Sprintf("must be empty when type is %s", failover.To.Type))
			}
		case AnyExcept, Only:
			if failover.To.Zones == nil || len(*failover.To.Zones) == 0 {
				verr.AddViolationAt(toZonesPath, fmt.Sprintf("must not be empty when type is %s", failover.To.Type))
			}
		default:
			verr.AddViolationAt(path.Field("to").Field("type"), "unrecognized type")
		}
	}

	verr.AddError("failoverThreshold", validateFailoverThreshold(crossZone.FailoverThreshold))

	return verr
}

func validateFailoverThreshold(failoverThreshold *FailoverThreshold) validators.ValidationError {
	var verr validators.ValidationError
	if failoverThreshold == nil {
		return verr
	}
	verr.Add(validators.ValidatePercentage(validators.RootedAt("percentage"), &failoverThreshold.Percentage, false))
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
		verr.AddError("leastRequest", validateLeastRequest(conf.LeastRequest))
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

func validateLeastRequest(conf *LeastRequest) validators.ValidationError {
	var verr validators.ValidationError
	if conf == nil {
		return verr
	}
	verr.Add(validators.ValidateIntOrStringGreaterOrEqualThan(validators.RootedAt("activeRequestBias"), conf.ActiveRequestBias, 0))
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
		case ConnectionType, SourceIPType:
			if policy.Connection == nil {
				verr.AddViolationAt(path.Field("connection"), validators.MustBeDefined)
			}
		}
	}
	return verr
}
