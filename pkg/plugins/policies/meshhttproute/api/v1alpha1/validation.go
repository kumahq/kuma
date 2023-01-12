package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

func (r *MeshHTTPRouteResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path.Field("to"), validateTos(r.Spec.To))
	return verr.OrNil()
}

func validateTop(targetRef common_api.TargetRef) validators.ValidationError {
	return matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.MeshSubset,
			common_api.MeshService,
			common_api.MeshServiceSubset,
		},
	})
}

func validateToRef(targetRef common_api.TargetRef) validators.ValidationError {
	return matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.MeshService,
		},
	})
}

func validateTos(tos []To) validators.ValidationError {
	var errs validators.ValidationError

	for i, to := range tos {
		path := validators.Root().Index(i)
		errs.AddErrorAt(path.Field("targetRef"), validateToRef(to.TargetRef))
		errs.AddErrorAt(path.Field("rules"), validateRules(to.Rules))
	}

	return errs
}

func validateRules(rules []Rule) validators.ValidationError {
	var errs validators.ValidationError

	for i, rule := range rules {
		path := validators.Root().Index(i)
		errs.AddErrorAt(path.Field("matches"), validateMatches(rule.Matches))
	}

	return errs
}

func validateMatches(matches []Match) validators.ValidationError {
	var errs validators.ValidationError

	for i, match := range matches {
		path := validators.Root().Index(i)
		errs.AddErrorAt(path.Field("path"), validatePath(match.Path))
		errs.AddErrorAt(path.Field("method"), validateMethod(match.Method))
	}

	return errs
}

func validatePath(match *PathMatch) validators.ValidationError {
	return validators.ValidationError{}
}

func validateMethod(match *Method) validators.ValidationError {
	return validators.ValidationError{}
}
