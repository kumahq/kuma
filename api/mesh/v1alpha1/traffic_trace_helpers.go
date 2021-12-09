package v1alpha1

import (
	"github.com/golang/protobuf/proto"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *TrafficTrace) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *TrafficTrace) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *TrafficTrace) DeepCopyInto(out *TrafficTrace) {
	proto.Merge(out, t)
}
func (t *TrafficTrace) DeepCopy() *TrafficTrace {
	if t == nil {
		return nil
	}
	out := new(TrafficTrace)
	t.DeepCopyInto(out)
	return out
}
