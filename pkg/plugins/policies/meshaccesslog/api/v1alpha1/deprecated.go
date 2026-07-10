package v1alpha1

import (
	"slices"

	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/jsonpatch/validators"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

func (t *MeshAccessLogResource) Deprecations() []string {
	var deprecations []string
	if len(pointer.Deref(t.Spec.From)) > 0 {
		deprecations = append(deprecations, "'from' field is deprecated, use 'rules' instead")
	}
	if hasInlineOtelEndpoint(t.Spec) {
		deprecations = append(deprecations, "openTelemetry.endpoint is deprecated, use openTelemetry.backendRef with a MeshOpenTelemetryBackend resource instead")
	}
	return append(deprecations, validators.TopLevelTargetRefDeprecations(t.Spec.TargetRef)...)
}

func hasInlineOtelEndpoint(spec *MeshAccessLog) bool {
	inlineEndpoint := func(b Backend) bool {
		return b.OpenTelemetry != nil && b.OpenTelemetry.Endpoint != "" && b.OpenTelemetry.BackendRef == nil
	}
	for _, to := range pointer.Deref(spec.To) {
		if slices.ContainsFunc(pointer.Deref(to.Default.Backends), inlineEndpoint) {
			return true
		}
	}
	for _, from := range pointer.Deref(spec.From) {
		if slices.ContainsFunc(pointer.Deref(from.Default.Backends), inlineEndpoint) {
			return true
		}
	}
	for _, rule := range pointer.Deref(spec.Rules) {
		if slices.ContainsFunc(pointer.Deref(rule.Default.Backends), inlineEndpoint) {
			return true
		}
	}
	return false
}
