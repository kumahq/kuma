package v1alpha1

import (
	"fmt"

	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/jsonpatch/validators"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

func deprecations(r *model.Res[*MeshRateLimit]) []string {
	var deprecations []string
	if len(pointer.Deref(r.Spec.From)) > 0 {
		deprecations = append(deprecations, "'from' field is deprecated, use 'rules' instead")
		for i, rule := range pointer.Deref(r.Spec.From) {
			if rule.Default.Local != nil && isStatusInvalid(*rule.Default.Local) {
				deprecations = append(deprecations, fmt.Sprintf("'spec.from[%d].default.local.http.requestRate.status' must be 400 or higher. Please update your configuration.", i))
			}
		}
	}
	if len(pointer.Deref(r.Spec.Rules)) > 0 {
		for i, rule := range pointer.Deref(r.Spec.Rules) {
			if rule.Default.Local != nil && isStatusInvalid(*rule.Default.Local) {
				deprecations = append(deprecations, fmt.Sprintf("'spec.rules[%d].default.local.http.requestRate.status' must be 400 or higher. Please update your configuration.", i))
			}
		}
	}
	if len(pointer.Deref(r.Spec.To)) > 0 {
		for i, rule := range pointer.Deref(r.Spec.To) {
			if rule.Default.Local != nil && isStatusInvalid(*rule.Default.Local) {
				deprecations = append(deprecations, fmt.Sprintf("'spec.to[%d].default.local.http.requestRate.status' must be 400 or higher. Please update your configuration.", i))
			}
		}
	}
	return append(deprecations, validators.TopLevelTargetRefDeprecations(r.Spec.TargetRef)...)
}

func isStatusInvalid(local Local) bool {
	if local.HTTP != nil &&
		local.HTTP.OnRateLimit != nil &&
		local.HTTP.OnRateLimit.Status != nil &&
		pointer.Deref(local.HTTP.OnRateLimit.Status) < 400 {
		return true
	}
	return false
}
