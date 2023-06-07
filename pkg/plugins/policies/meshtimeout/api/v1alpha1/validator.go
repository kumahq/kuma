package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/core/matchers/validators"
)

func (r *MeshTimeoutResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	if len(r.Spec.To) == 0 && len(r.Spec.From) == 0 {
		verr.AddViolationAt(path, "at least one of 'from', 'to' has to be defined")
	}
	verr.AddErrorAt(path, validateFrom(r.Spec.From))
	verr.AddErrorAt(path, validateTo(r.Spec.To))
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
		verr.AddErrorAt(path.Field("targetRef"), matcher_validators.ValidateTargetRef(fromItem.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
			},
		}))

		defaultField := path.Field("default")
		verr.Add(validateDefault(defaultField, fromItem.Default))
	}
	return verr
}

func validateTo(to []To) validators.ValidationError {
	var verr validators.ValidationError
	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), matcher_validators.ValidateTargetRef(toItem.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshService,
			},
		}))

		defaultField := path.Field("default")
		verr.Add(validateDefault(defaultField, toItem.Default))
	}
	return verr
}

func validateDefault(path validators.PathBuilder, conf Conf) validators.ValidationError {
	var verr validators.ValidationError

	if conf.ConnectionTimeout == nil && conf.IdleTimeout == nil && conf.Http == nil {
		verr.AddViolationAt(path, "at least one timeout should be configured")
		return verr
	}

	verr.Add(validators.ValidateDurationGreaterThanZeroOrNil(path.Field("connectionTimeout"), conf.ConnectionTimeout))
	verr.Add(validators.ValidateDurationNotNegativeOrNil(path.Field("idleTimeout"), conf.IdleTimeout))

	verr.Add(validateHttp(path.Field("http"), conf.Http))
	return verr
}

func validateHttp(path validators.PathBuilder, http *Http) validators.ValidationError {
	var verr validators.ValidationError
	if http == nil {
		return verr
	}

	if http.RequestTimeout == nil && http.StreamIdleTimeout == nil && http.MaxStreamDuration == nil && http.MaxConnectionDuration == nil {
		verr.AddViolationAt(path, "at least one timeout in this section should be configured")
	}

	verr.Add(validators.ValidateDurationNotNegativeOrNil(path.Field("requestTimeout"), http.RequestTimeout))
	verr.Add(validators.ValidateDurationNotNegativeOrNil(path.Field("streamIdleTimeout"), http.StreamIdleTimeout))
	verr.Add(validators.ValidateDurationNotNegativeOrNil(path.Field("maxStreamDuration"), http.MaxStreamDuration))
	verr.Add(validators.ValidateDurationNotNegativeOrNil(path.Field("maxConnectionDuration"), http.MaxConnectionDuration))

	return verr
}
