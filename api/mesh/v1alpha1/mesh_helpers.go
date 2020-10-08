package v1alpha1

func (m *Mesh) IsPassthrough() bool {
	if m.GetNetworking() == nil || m.GetNetworking().GetOutbound() == nil || m.GetNetworking().GetOutbound().GetPassthrough() == nil {
		return true
	}
	return m.GetNetworking().GetOutbound().GetPassthrough().Value
}
