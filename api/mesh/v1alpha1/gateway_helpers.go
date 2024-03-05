package v1alpha1

const WildcardHostname = "*"

func (g *MeshGateway) IsCrossMesh() bool {
	for _, l := range g.GetConf().GetListeners() {
		if l.CrossMesh {
			return true
		}
	}
	return false
}

// GetNonEmptyHostname returns "*" if the hostname isn't set.
func (l *MeshGateway_Listener) GetNonEmptyHostname() string {
	if h := l.GetHostname(); h != "" {
		return h
	}
	return WildcardHostname
}
