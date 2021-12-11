package v1alpha1

import (
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *TrafficRoute) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *TrafficRoute) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *TrafficRoute) DeepCopyInto(out *TrafficRoute) {
	util_proto.Merge(out, t)
}
func (t *TrafficRoute) DeepCopy() *TrafficRoute {
	if t == nil {
		return nil
	}
	out := new(TrafficRoute)
	t.DeepCopyInto(out)
	return out
}
