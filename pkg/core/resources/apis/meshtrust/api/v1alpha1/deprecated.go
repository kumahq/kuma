package v1alpha1

func (r *MeshTrustResource) Deprecations() []string {
	var deprecations []string
	if r.Spec != nil && r.Spec.Origin != nil {
		deprecations = append(deprecations, "setting spec.origin field is deprecated")
	}
	return deprecations
}
