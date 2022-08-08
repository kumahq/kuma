package v1alpha1

import (
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

const WildcardHostname = "*"

func (m *MeshGateway) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *MeshGateway) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *MeshGateway) DeepCopyInto(out *MeshGateway) {
	util_proto.Merge(out, t)
}
func (t *MeshGateway) DeepCopy() *MeshGateway {
	if t == nil {
		return nil
	}
	out := new(MeshGateway)
	t.DeepCopyInto(out)
	return out
}

func (g *MeshGateway) IsCrossMesh() bool {
	for _, l := range g.GetConf().GetListeners() {
		if l.CrossMesh {
			return true
		}
	}
	return false
}
