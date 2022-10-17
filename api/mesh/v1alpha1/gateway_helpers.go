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
