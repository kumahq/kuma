// +kubebuilder:object:generate=true
package v1alpha1

// MeshZoneAddress holds the public address and port for a mesh-scoped zone ingress proxy.
// +kuma:policy:is_policy=false
// +kuma:policy:kds_flags=model.ZoneToGlobalFlag | model.SyncedAcrossZonesFlag
// +kuma:policy:short_name=mza
// +kuma:policy:allowed_on_system_namespace_only=false
type MeshZoneAddress struct {
	// Address is the publicly reachable address of the zone ingress.
	Address string `json:"address"`
	// Port is the publicly reachable port of the zone ingress.
	Port uint32 `json:"port"`
}
