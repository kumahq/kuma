package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (r *MeshTrafficPermissionResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), r.validateTop(r.Spec.TargetRef, inbound.AffectsInbounds(r.Spec)))
	if len(pointer.Deref(r.Spec.From)) > 0 && len(pointer.Deref(r.Spec.Rules)) > 0 {
		verr.AddViolationAt(path, "field 'from' must be empty when 'rules' is defined")
	}
	if len(pointer.Deref(r.Spec.From)) == 0 && len(pointer.Deref(r.Spec.Rules)) == 0 {
		verr.AddViolationAt(path, "at least one of 'from' or 'rules' has to be defined")
	}
	verr.AddErrorAt(path, validateRules(pointer.Deref(r.Spec.Rules)))
	verr.AddErrorAt(path, validateFrom(pointer.Deref(r.Spec.From)))
	return verr.OrNil()
}

func (r *MeshTrafficPermissionResource) validateTop(targetRef *common_api.TargetRef, isInboundPolicy bool) validators.ValidationError {
	if targetRef == nil {
		return validators.ValidationError{}
	}
	targetRefErr := mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.MeshSubset,
			common_api.MeshService,
			common_api.MeshServiceSubset,
			common_api.Dataplane,
		},
		IsInboundPolicy: isInboundPolicy,
	})
	return targetRefErr
}

func validateRules(rules []Rule) validators.ValidationError {
	var verr validators.ValidationError
	for idx, rule := range rules {
		path := validators.RootedAt("rules").Index(idx)
		if len(pointer.Deref(rule.Default.Deny)) == 0 && len(pointer.Deref(rule.Default.Allow)) == 0 && len(pointer.Deref(rule.Default.AllowWithShadowDeny)) == 0 {
			verr.AddViolationAt(path, "at least one of 'allow', 'allowWithShadowDeny', 'deny' has to be defined")
		}
		verr.AddErrorAt(path, validateMatches("allow", pointer.Deref(rule.Default.Allow)))
		verr.AddErrorAt(path, validateMatches("allowWithShadowDeny", pointer.Deref(rule.Default.AllowWithShadowDeny)))
		verr.AddErrorAt(path, validateMatches("deny", pointer.Deref(rule.Default.Deny)))
	}
	return verr
}

func validateMatches(field string, matches []common_api.Match) validators.ValidationError {
	var verr validators.ValidationError
	for idx, match := range matches {
		path := validators.RootedAt(field).Index(idx)
		verr.AddErrorAt(path, mesh.ValidateMatch(match))
	}
	return verr
}

func validateFrom(from []From) validators.ValidationError {
	var verr validators.ValidationError
	for idx, fromItem := range from {
		path := validators.RootedAt("from").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), mesh.ValidateTargetRef(fromItem.GetTargetRef(), &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshSubset,
				common_api.MeshService,
				common_api.MeshServiceSubset,
			},
		}))

		defaultField := path.Field("default")
		verr.AddErrorAt(defaultField, validateDefault(fromItem.Default))
	}
	return verr
}

func validateDefault(conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	if len(pointer.Deref(conf.Action)) == 0 {
		verr.AddViolation("action", validators.MustBeDefined)
	}
	return verr
}
