package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

func (r *MeshFaultInjectionResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path, validateFrom(r.Spec.From))
	return verr.OrNil()
}

func validateTop(targetRef common_api.TargetRef) validators.ValidationError {
	targetRefErr := matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.MeshSubset,
			common_api.MeshService,
			common_api.MeshServiceSubset,
		},
	})
	return targetRefErr
}

func validateFrom(from []From) validators.ValidationError {
	var verr validators.ValidationError
	for idx, fromItem := range from {
		path := validators.RootedAt("from").Index(idx)
		defaultField := path.Field("default")
		verr.AddErrorAt(path.Field("targetRef"), matcher_validators.ValidateTargetRef(fromItem.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
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

func validateDefault(path validators.PathBuilder, conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	path = path.Field("http")
	for idx, fault := range conf.Http {
		if fault.Abort != nil {
			path := path.Field("abort").Index(idx)
			verr.Add(validators.ValidateStatusCode(path.Field("httpStatus"), fault.Abort.HttpStatus))
			verr.Add(validators.ValidatePercentage(path.Field("percentage"), &fault.Abort.Percentage))
		}
		if fault.Delay != nil {
			path := path.Field("delay").Index(idx)
			verr.Add(validators.ValidateDurationNotNegative(path.Field("value"), &fault.Delay.Value))
			verr.Add(validators.ValidatePercentage(path.Field("percentage"), &fault.Delay.Percentage))
		}
		if fault.ResponseBandwidth != nil {
			path := path.Field("responseBandwidth").Index(idx)
			verr.Add(validators.ValidateBandwidth(path.Field("responseBandwidth"), fault.ResponseBandwidth.Limit))
			verr.Add(validators.ValidatePercentage(path.Field("percentage"), &fault.ResponseBandwidth.Percentage))
		}
	}
	return verr
}
