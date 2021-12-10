package v1alpha1

import (
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *TrafficLog) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *TrafficLog) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *TrafficLog) DeepCopyInto(out *TrafficLog) {
	util_proto.Merge(out, t)
}
func (t *TrafficLog) DeepCopy() *TrafficLog {
	if t == nil {
		return nil
	}
	out := new(TrafficLog)
	t.DeepCopyInto(out)
	return out
}
