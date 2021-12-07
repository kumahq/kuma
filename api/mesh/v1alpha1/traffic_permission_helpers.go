package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	proto_util "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *TrafficPermission) UnmarshalJSON(data []byte) error {
	return proto_util.FromJSON(data, m)
}

func (m *TrafficPermission) MarshalJSON() ([]byte, error) {
	return proto_util.ToJSON(m)
}
func (t *TrafficPermission) DeepCopyInto(out *TrafficPermission) {
	proto.Merge(out, t)
}
func (t *TrafficPermission) DeepCopy() *TrafficPermission {
	if t == nil {
		return nil
	}
	out := new(TrafficPermission)
	t.DeepCopyInto(out)
	return out
}
