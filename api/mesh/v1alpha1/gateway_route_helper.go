package v1alpha1

import (
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *MeshGatewayRoute) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *MeshGatewayRoute) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *MeshGatewayRoute) DeepCopyInto(out *MeshGatewayRoute) {
	util_proto.Merge(out, t)
}
func (t *MeshGatewayRoute) DeepCopy() *MeshGatewayRoute {
	if t == nil {
		return nil
	}
	out := new(MeshGatewayRoute)
	t.DeepCopyInto(out)
	return out
}
