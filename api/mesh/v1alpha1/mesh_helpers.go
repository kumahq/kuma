package v1alpha1

func (m *Mesh) IsPassthrough() bool {
	passthrough := m.GetNetworking().GetOutbound().GetPassthrough()
	if passthrough == nil {
		return true
	}
	return passthrough.GetValue()
}

func (m *Mesh) MeshServicesMode() Mesh_MeshServices_Mode {
	if m.MeshServices == nil {
		return Mesh_MeshServices_Disabled
	}
	return m.MeshServices.Mode
}
