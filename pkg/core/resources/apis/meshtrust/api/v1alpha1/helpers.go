package v1alpha1

func GetAllTrustDomains(meshTrusts MeshTrustResourceList) map[string]struct{} {
	trustDomains := map[string]struct{}{}
	for _, meshTrust := range meshTrusts.Items {
		trustDomains[meshTrust.Spec.TrustDomain] = struct{}{}
	}
	return trustDomains
}

// MigrateOriginToStatus migrates the origin field from spec to status
// for backward compatibility during the transition period.
// Returns true if migration was performed, false otherwise.
func (r *MeshTrustResource) MigrateOriginToStatus() bool {
	if r.Spec == nil || r.Spec.Origin == nil {
		return false
	}
	if r.Status == nil {
		r.Status = &MeshTrustStatus{}
	}
	if r.Status.Origin == nil {
		r.Status.Origin = r.Spec.Origin
		return true
	}
	return false
}
