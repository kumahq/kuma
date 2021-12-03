package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	proto_util "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *Mesh) IsPassthrough() bool {
	passthrough := m.GetNetworking().GetOutbound().GetPassthrough()
	if passthrough == nil {
		return true
	}
	return passthrough.GetValue()
}

func (m *Mesh) UnmarshalJSON(data []byte) error {
	return proto_util.FromJSON(data, m)
}

func (m *Mesh) MarshalJSON() ([]byte, error) {
	return proto_util.ToJSON(m)
}
func (t *Mesh) DeepCopyInto(out *Mesh) {
	proto.Merge(out, t)
}
