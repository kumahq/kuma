package v1alpha1

import (
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func (m *HealthCheck) UnmarshalJSON(data []byte) error {
	return util_proto.FromJSON(data, m)
}

func (m *HealthCheck) MarshalJSON() ([]byte, error) {
	return util_proto.ToJSON(m)
}
func (t *HealthCheck) DeepCopyInto(out *HealthCheck) {
	util_proto.Merge(out, t)
}
func (t *HealthCheck) DeepCopy() *HealthCheck {
	if t == nil {
		return nil
	}
	out := new(HealthCheck)
	t.DeepCopyInto(out)
	return out
}
