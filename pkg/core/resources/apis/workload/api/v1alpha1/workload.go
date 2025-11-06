// +kubebuilder:object:generate=true
package v1alpha1

// Workload
// +kuma:policy:is_policy=false
// +kuma:policy:has_status=true
// +kuma:policy:kds_flags=model.ZoneToGlobalFlag
// +kuma:policy:short_name=wl
type Workload struct{}
type WorkloadStatus struct {
	// DataplaneProxies defines statistics of data plane proxies that are part of this workload
	DataplaneProxies DataplaneProxies `json:"dataplaneProxies,omitempty"`
}

type DataplaneProxies struct {
	// Connected defines number of connected data plane proxies
	Connected int32 `json:"connected"`
	// Healthy defines number of healthy data plane proxies for this workload
	Healthy int32 `json:"healthy"`
	// Total defines total number of data plane proxies for this workload
	Total int32 `json:"total"`
}
