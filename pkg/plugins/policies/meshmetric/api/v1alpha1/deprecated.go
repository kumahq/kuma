package v1alpha1

import (
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/jsonpatch/validators"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

func (t *MeshMetricResource) Deprecations() []string {
	var deprecations []string
	for _, backend := range pointer.Deref(t.Spec.Default.Backends) {
		if backend.OpenTelemetry != nil && backend.OpenTelemetry.Endpoint != "" && backend.OpenTelemetry.BackendRef == nil {
			deprecations = append(deprecations, "openTelemetry.endpoint is deprecated, use openTelemetry.backendRef with a MeshOpenTelemetryBackend resource instead")
			break
		}
	}
	return append(deprecations, validators.TopLevelTargetRefDeprecations(t.Spec.TargetRef)...)
}
