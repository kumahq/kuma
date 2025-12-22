package v1alpha1

func (t *MeshTrustResource) Deprecations() []string {
	var deprecations []string
	if t.Spec != nil && t.Spec.Origin != nil {
		deprecations = append(deprecations, "setting spec.origin field is deprecated")
	}
	return deprecations
}
