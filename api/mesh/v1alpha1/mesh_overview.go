package v1alpha1

// MeshOverview defines the projected state of a Mesh. It is assembled for the API and
// never stored.
type MeshOverview struct {
	Mesh *Mesh `json:"mesh,omitempty"`
	// MeshInsight carries the observed state of the mesh.
	MeshInsight *MeshInsight `json:"meshInsight,omitempty"`
}

func (o *MeshOverview) GetMesh() *Mesh {
	if o == nil {
		return nil
	}
	return o.Mesh
}

func (o *MeshOverview) GetMeshInsight() *MeshInsight {
	if o == nil {
		return nil
	}
	return o.MeshInsight
}
