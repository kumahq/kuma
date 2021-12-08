package v1alpha1

import (
	"github.com/golang/protobuf/proto"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *TrafficPermission) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *TrafficPermission) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
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
