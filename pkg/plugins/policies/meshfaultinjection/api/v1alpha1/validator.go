package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (r *MeshFaultInjectionResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), r.validateTop(r.Spec.TargetRef))
	topLevel := pointer.DerefOr(r.Spec.TargetRef, common_api.TargetRef{Kind: common_api.Mesh})
	verr.AddErrorAt(path, validateFrom(topLevel, r.Spec.From))
	verr.AddErrorAt(path, validateTo(topLevel, r.Spec.To))
	return verr.OrNil()
}

func (r *MeshFaultInjectionResource) validateTop(targetRef *common_api.TargetRef) validators.ValidationError {
	if targetRef == nil {
		return validators.ValidationError{}
	}
	switch core_model.PolicyRole(r.GetMeta()) {
	case mesh_proto.SystemPolicyRole:
		return mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshSubset,
				common_api.MeshGateway,
				common_api.MeshService,
				common_api.MeshServiceSubset,
			},
			GatewayListenerTagsAllowed: true,
		})
	default:
		return mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshSubset,
				common_api.MeshService,
				common_api.MeshServiceSubset,
			},
		})
	}
}

func validateFrom(topTargetRef common_api.TargetRef, from []From) validators.ValidationError {
	var verr validators.ValidationError
	if common_api.IncludesGateways(topTargetRef) && len(from) != 0 {
		verr.AddViolationAt(validators.RootedAt("from"), validators.MustNotBeDefined)
		return verr
	}
	for idx, fromItem := range from {
		path := validators.RootedAt("from").Index(idx)
		defaultField := path.Field("default")
		verr.AddErrorAt(path.Field("targetRef"), mesh.ValidateTargetRef(fromItem.GetTargetRef(), &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshSubset,
				common_api.MeshServiceSubset,
				common_api.MeshService,
			},
		}))
		verr.Add(validateDefault(defaultField, fromItem.Default))
	}
	return verr
}

func validateTo(topTargetRef common_api.TargetRef, to []To) validators.ValidationError {
	var verr validators.ValidationError
	if !common_api.IncludesGateways(topTargetRef) && len(to) != 0 {
		verr.AddViolationAt(validators.RootedAt("to"), validators.MustNotBeDefined)
		return verr
	}
	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)
		defaultField := path.Field("default")
		verr.AddErrorAt(path.Field("targetRef"), mesh.ValidateTargetRef(toItem.GetTargetRef(), &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
			},
		}))
		verr.Add(validateDefault(defaultField, toItem.Default))
	}
	return verr
}

func validateDefault(path validators.PathBuilder, conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	path = path.Field("http")
	cfg := pointer.Deref(conf.Http)
	for idx := range cfg {
		fault := cfg[idx]
		if fault.Abort != nil {
			path := path.Field("abort").Index(idx)
			verr.Add(validators.ValidateStatusCode(path.Field("httpStatus"), fault.Abort.HttpStatus))
			verr.Add(validators.ValidatePercentage(path.Field("percentage"), &fault.Abort.Percentage, true))
		}
		if fault.Delay != nil {
			path := path.Field("delay").Index(idx)
			verr.Add(validators.ValidateDurationNotNegative(path.Field("value"), &fault.Delay.Value))
			verr.Add(validators.ValidatePercentage(path.Field("percentage"), &fault.Delay.Percentage, true))
		}
		if fault.ResponseBandwidth != nil {
			path := path.Field("responseBandwidth").Index(idx)
			verr.Add(validators.ValidateBandwidth(path.Field("responseBandwidth"), fault.ResponseBandwidth.Limit))
			verr.Add(validators.ValidatePercentage(path.Field("percentage"), &fault.ResponseBandwidth.Percentage, true))
		}
	}
	return verr
}
