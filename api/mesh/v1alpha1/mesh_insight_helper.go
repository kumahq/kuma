package v1alpha1

import (
	"github.com/golang/protobuf/proto"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *MeshInsight) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *MeshInsight) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
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
