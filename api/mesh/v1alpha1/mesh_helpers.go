package v1alpha1

import (
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *Mesh) IsPassthrough() bool {
	passthrough := m.GetNetworking().GetOutbound().GetPassthrough()
	if passthrough == nil {
		return true
	}
	return passthrough.GetValue()
}

func (m *Mesh) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *Mesh) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *Mesh) DeepCopyInto(out *Mesh) {
	util_proto.Merge(out, t)
}

func (t *Mesh) DeepCopy() *Mesh {
	if t == nil {
		return nil
	}
	out := new(Mesh)
	t.DeepCopyInto(out)
	return out
}
