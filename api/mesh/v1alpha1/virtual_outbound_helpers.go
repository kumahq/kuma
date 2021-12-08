package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *VirtualOutbound) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *VirtualOutbound) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *VirtualOutbound) DeepCopyInto(out *VirtualOutbound) {
	proto.Merge(out, t)
}
func (t *VirtualOutbound) DeepCopy() *VirtualOutbound {
	if t == nil {
		return nil
	}
	out := new(VirtualOutbound)
	t.DeepCopyInto(out)
	return out
}
