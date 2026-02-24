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
	return slices.ContainsFunc(allBackends(spec), func(b Backend) bool {
		return b.OpenTelemetry != nil && b.OpenTelemetry.Endpoint != "" && b.OpenTelemetry.BackendRef == nil
	})
}

func allBackends(spec *MeshAccessLog) []Backend {
	var backends []Backend
	collectFrom := func(conf Conf) {
		backends = append(backends, pointer.Deref(conf.Backends)...)
	}
	for _, to := range pointer.Deref(spec.To) {
		collectFrom(to.Default)
	}
	for _, from := range pointer.Deref(spec.From) {
		collectFrom(from.Default)
	}
	for _, rule := range pointer.Deref(spec.Rules) {
		collectFrom(rule.Default)
	}
	return backends
}
