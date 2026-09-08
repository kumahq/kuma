// +kubebuilder:object:generate=true
package v1alpha1

// Zone defines the Zone configuration used at the Global Control Plane within a
// distributed deployment. Enabled is a pointer because an unset value and an explicit
// false mean different things: protobuf omitted an unset wrapper entirely, and an older
// control plane still distinguishes the two.
// +kuma:policy:is_policy=false
// +kuma:policy:scope=Global
// +kuma:policy:cluster_scoped_k8s=true
// +kuma:policy:policy_matching_exempt=true
// +kuma:policy:opaque_k8s_spec=true
// +kuma:policy:plugin_originated=false
// +kuma:policy:ws_name=zone
// +kuma:policy:short_name=z
// +kuma:policy:has_insights=true
// +kuma:policy:insight_package=github.com/kumahq/kuma/v3/pkg/core/resources/apis/zoneinsight/api/v1alpha1
// +kuma:policy:kds_flags=model.ProvidedByGlobalFlag | model.ProvidedByZoneFlag
type Zone struct {
	Enabled *bool `json:"enabled,omitempty"`
}

func (z *Zone) IsEnabled() bool {
	return z.GetEnabled()
}

func (z *Zone) GetEnabled() bool {
	if z == nil || z.Enabled == nil {
		return true
	}
	return *z.Enabled
}
