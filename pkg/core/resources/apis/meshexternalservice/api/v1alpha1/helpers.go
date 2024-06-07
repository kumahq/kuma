package v1alpha1

func (m *MeshExternalServiceResource) DestinationName(port uint32) string {
	return m.GetMeta().GetName()
}
