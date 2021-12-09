package v1alpha1

import (
	"github.com/golang/protobuf/proto"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *ZoneIngress) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *ZoneIngress) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *ZoneIngress) DeepCopyInto(out *ZoneIngress) {
	proto.Merge(out, t)
}
func (t *ZoneIngress) DeepCopy() *ZoneIngress {
	if t == nil {
		return nil
	}
	out := new(ZoneIngress)
	t.DeepCopyInto(out)
	return out
}
