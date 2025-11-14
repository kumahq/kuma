// +kubebuilder:object:generate=true
package v1alpha1

// Workload represents a logical grouping of data plane proxies in the mesh, providing visibility into their operational status. It tracks statistics about the data plane proxies that belong to a workload, including the number of connected, healthy, and total proxies, enabling monitoring and health assessment of your workload deployments. Workloads is also the primary way data-planes are grouped together in metrics and traces.
// +kuma:policy:is_policy=false
// +kuma:policy:has_status=true
// +kuma:policy:kds_flags=model.ZoneToGlobalFlag
// +kuma:policy:short_name=wl
type (
	Workload       struct{}
	WorkloadStatus struct {
		// DataplaneProxies defines statistics of data plane proxies that are part of this workload
		DataplaneProxies DataplaneProxies `json:"dataplaneProxies,omitempty"`
	}
)

type DataplaneProxies struct {
	// Connected defines number of connected data plane proxies
	Connected int32 `json:"connected"`
	// Healthy defines number of healthy data plane proxies for this workload
	Healthy int32 `json:"healthy"`
	// Total defines total number of data plane proxies for this workload
	Total int32 `json:"total"`
}
