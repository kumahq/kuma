package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	proto_util "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *ZoneIngress) UnmarshalJSON(data []byte) error {
	return proto_util.FromJSON(data, m)
}

func (m *ZoneIngress) MarshalJSON() ([]byte, error) {
	return proto_util.ToJSON(m)
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
