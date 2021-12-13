package v1alpha1

import (
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *ServiceInsight) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *ServiceInsight) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *ServiceInsight) DeepCopyInto(out *ServiceInsight) {
	util_proto.Merge(out, t)
}
func (t *ServiceInsight) DeepCopy() *ServiceInsight {
	if t == nil {
		return nil
	}
	out := new(ServiceInsight)
	t.DeepCopyInto(out)
	return out
}
