package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	proto_util "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *MeshInsight) UnmarshalJSON(data []byte) error {
	return proto_util.FromJSON(data, m)
}

func (m *MeshInsight) MarshalJSON() ([]byte, error) {
	return proto_util.ToJSON(m)
}
func (t *MeshInsight) DeepCopyInto(out *MeshInsight) {
	proto.Merge(out, t)
}
func (t *MeshInsight) DeepCopy() *MeshInsight {
	if t == nil {
		return nil
	}
	out := new(MeshInsight)
	t.DeepCopyInto(out)
	return out
}
