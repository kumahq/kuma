package v1alpha1

import (
	"fmt"

	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/jsonpatch/validators"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func (t *MeshRateLimitResource) Deprecations() []string {
	var deprecations []string
	if len(pointer.Deref(t.Spec.Rules)) > 0 {
		for i, rule := range pointer.Deref(t.Spec.Rules) {
			if rule.Default.Local != nil && isStatusInvalid(*rule.Default.Local) {
				deprecations = append(deprecations, fmt.Sprintf("'spec.rules[%d].default.local.http.requestRate.status' must be 400 or higher. Please update your configuration.", i))
			}
		}
	}
	if len(pointer.Deref(t.Spec.To)) > 0 {
		for i, rule := range pointer.Deref(t.Spec.To) {
			if rule.Default.Local != nil && isStatusInvalid(*rule.Default.Local) {
				deprecations = append(deprecations, fmt.Sprintf("'spec.to[%d].default.local.http.requestRate.status' must be 400 or higher. Please update your configuration.", i))
			}
		}
	}
	return append(deprecations, validators.TopLevelTargetRefDeprecations(t.Spec.TargetRef)...)
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
