package v1alpha1

// Mesh defines configuration of a single mesh.
type Mesh struct {
	// List of policies to skip creating by default when the mesh is created,
	// for example TrafficPermission or MeshRetry. An '*' skips all policies.
	SkipCreatingInitialPolicies []string `json:"skipCreatingInitialPolicies,omitempty"`
}

func (m *Mesh) GetSkipCreatingInitialPolicies() []string {
	if m == nil {
		return nil
	}
	return m.SkipCreatingInitialPolicies
}
