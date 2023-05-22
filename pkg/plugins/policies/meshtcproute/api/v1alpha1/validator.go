package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

func (r *MeshTCPRouteResource) validate() error {
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
	return matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.MeshSubset,
			common_api.MeshService,
			common_api.MeshServiceSubset,
		},
	})
}

func validateTo(to []To) validators.ValidationError {
	var verr validators.ValidationError

	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), matcher_validators.ValidateTargetRef(toItem.TargetRef, &matcher_validators.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.MeshService,
			},
		}))
		verr.AddErrorAt(path.Field("rules"), validateRules(toItem.Rules))
	}

	return verr
}

func validateRules(rules []Rule) validators.ValidationError {
	var verr validators.ValidationError

	if len(rules) != 1 {
		// At this point there is no plan to introduce address matching
		// capabilities for `MeshTCPRoute` in foreseeable future. We try to be
		// as close with structures of our policies to the Gateway API
		// as possible. It means, that even if Gateway API currently doesn't
		// have plans to support this kind of matching as well (ref.
		// Kubernetes Gateway API GEP-735: TCP and UDP addresses matching -
		// https://gateway-api.sigs.k8s.io/geps/gep-735/), its structures
		// are ready to potentially support it.
		//
		// As a result every element of the route destination section of
		// the policy configuration (`spec.to[]`) contains a `rules` property.
		// This property is a list of elements, which potentially will allow
		// to specify `match` configuration.
		//
		// Without specifying `match`es, it would be nonsensical to accept more
		// `rules.`
		verr.AddViolationAt(validators.Root(), "needs exactly one item")
		return verr
	}

	for i, rule := range rules {
		path := validators.Root().Index(i)
		verr.AddErrorAt(path.Field("default").Field("backendRefs"), validateBackendRefs(rule.Default.BackendRefs))
	}

	return verr
}

func validateBackendRefs(backendRefs *[]BackendRef) validators.ValidationError {
	var verr validators.ValidationError

	if backendRefs == nil {
		return verr
	}

	for i, backendRef := range *backendRefs {
		verr.AddErrorAt(
			validators.Root().Index(i),
			matcher_validators.ValidateTargetRef(backendRef.TargetRef, &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
					common_api.MeshServiceSubset,
				},
			}),
		)
	}

	return verr
}
