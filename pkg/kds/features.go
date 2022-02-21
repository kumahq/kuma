package kds

// Features is a set of available features for the control plane.
// If by any chance we get into a situation that we need to execute a logic conditionally on capabilities of control plane,
// instead of defining conditions on version which is fragile, we can define a condition based on features.
type Features map[string]bool

func (f Features) HasFeature(feature string) bool {
	return f[feature]
}

// FeatureZoneToken means that the zone control plane can handle incoming Zone Token from global control plane.
const FeatureZoneToken string = "zone-token"
